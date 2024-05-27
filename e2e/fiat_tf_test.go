package e2e

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

func TestFiatTFMint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Mint while TF is paused
	// EXPECTED: Request fails; amount not minted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "receiver-1", math.OneInt(), noble)
	receiver1 := w[0]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	showMinterPreMint, err := showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")

	preMintAllowance := showMinterPreMint.Minters.Allowance.Amount

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "minting is paused")

	bal, err := noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMinterPostMint, err := showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: nw.fiatTfRoles.Minter.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: preMintAllowance,
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterPostMint.Minters)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Mint from non minter
	// EXPECTED: Request fails; amount not minted

	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "you are not a minter")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// ACTION: Mint from blacklisted minter
	// EXPECTED: Request fails; amount not minted

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "minter address is blacklisted")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMintersRes, err := showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")

	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	// ACTION: Mint to blacklisted account
	// EXPECTED: Request fails; amount not minted

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver1)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "receiver address is blacklisted")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMintersRes, err = showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")

	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver1)

	// ACTION: Mint an amount that exceeds the minters allowance
	// EXPECTED: Request fails; amount not minted

	exceedAllowance := preMintAllowance.Add(math.NewInt(99))
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", exceedAllowance.Int64()))
	require.ErrorContains(t, err, "minting amount is greater than the allowance")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMintersRes, err = showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")
	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	// ACTION: Successfully mint into an account
	// EXPECTED: Success

	mintAmount := int64(3)
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.Equal(math.NewInt(mintAmount)))

	showMintersRes, err = showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters.Minters.Allowance = sdktypes.Coin{
		Denom:  "uusdc",
		Amount: preMintAllowance.Sub(math.NewInt(mintAmount)),
	}

	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)
}

func TestFiatTFBurn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// setup - mint into minter's wallet
	mintAmount := int64(5)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nw.fiatTfRoles.Minter.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err := noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64())

	// ACTION: Burn while TF is paused
	// EXPECTED: Request fails; amount not burned

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	burnAmount := int64(1)
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "burning is paused")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Burn from non minter account
	// EXPECTED: Request fails; amount not burned

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.NewInt(burnAmount), noble)
	alice := w[0]

	// mint into Alice's account to give her a balance to burn
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err = noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.Equal(math.NewInt(mintAmount)))

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "you are not a minter")

	bal, err = noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	// ACTION: Burn from a blacklisted minter account
	// EXPECTED: Request fails; amount not burned

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "minter address is blacklisted")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	// ACTION: Burn amount greater than the minters balance
	// EXPECTED: Request fails; amount not burned

	exceedAllowance := bal.Add(math.NewInt(99))
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", exceedAllowance.Int64()))
	require.ErrorContains(t, err, "insufficient funds")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	// ACTION: Successfully burn tokens
	// EXPECTED: Success; amount burned and Minters balance is decreased

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	expectedAmount := mintAmount - burnAmount
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, expectedAmount, bal.Int64())
}
