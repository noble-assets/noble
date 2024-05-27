package e2e

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

func TestFiatTFAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Send non TF token but pay fee in TF token while the TF is paused
	// EXPECTED: Request fails; TF is paused

	originalAmount := math.OneInt()
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", originalAmount, noble, noble)
	alice := w[0] // 1ustake
	bob := w[1]   // 1ustake

	mintAmount := 100
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err)

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	sendAmount := 1
	uusdcFee := 5
	_, err = val.ExecTx(ctx, alice.KeyName(), "bank", "send", alice.KeyName(), bob.FormattedAddress(), fmt.Sprintf("%dustake", sendAmount), "--fees", fmt.Sprintf("%duusdc", uusdcFee))
	require.ErrorContains(t, err, "the chain is paused")

	bal, err := noble.GetBalance(ctx, alice.FormattedAddress(), "ustake")
	require.NoError(t, err)
	require.Equal(t, originalAmount, bal)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Send non TF token but pay fee in TF token while the sender is blacklisted
	// EXPECTED: Request fails

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, alice)

	_, err = val.ExecTx(ctx, alice.KeyName(), "bank", "send", alice.KeyName(), bob.FormattedAddress(), fmt.Sprintf("%dustake", sendAmount), "--fees", fmt.Sprintf("%duusdc", uusdcFee))
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens: unauthorized", alice.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, alice.FormattedAddress(), "ustake")
	require.NoError(t, err)
	require.Equal(t, originalAmount, bal)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, alice)

	// ACTION: Successfully send non TF token but pay fee in TF token
	// EXPECTED: Success; Fee withdrawn from users balance

	_, err = val.ExecTx(ctx, alice.KeyName(), "bank", "send", alice.KeyName(), bob.FormattedAddress(), fmt.Sprintf("%dustake", sendAmount), "--fees", fmt.Sprintf("%duusdc", uusdcFee))
	require.NoError(t, err)

	bobBalStake, err := noble.GetBalance(ctx, bob.FormattedAddress(), "ustake")
	require.NoError(t, err)
	require.Equal(t, originalAmount.Add(math.NewInt(int64(sendAmount))), bobBalStake)

	aliceBalUusdc, err := noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.EqualValues(t, mintAmount-uusdcFee, aliceBalUusdc.Int64())
}

func TestFiatTFAuthzGrant(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Grant an authz SEND using a TF token while TF is paused
	// EXPECTED: Request fails

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter1 := w[0]
	grantee1 := w[1]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err := val.AuthzGrant(ctx, granter1, grantee1.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.ErrorContains(t, err, "can not perform token authorizations: the chain is paused")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Grant an authz SEND using a TF token to a grantee who is blacklisted
	// EXPECTED: Success;

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter2 := w[0]
	grantee2 := w[1]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, grantee2)

	res, err := val.AuthzGrant(ctx, granter2, grantee2.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	// ACTION: Grant an authz SEND using a TF token from a granter who is blacklisted
	// EXPECTED: Success;

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter3 := w[0]
	grantee3 := w[1]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, granter3)

	res, err = val.AuthzGrant(ctx, granter3, grantee3.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)
}

func TestFiatTFAuthzSend(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// setup
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble, noble)
	granter := w[0]
	grantee := w[1]
	receiver := w[2]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", granter.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	res, err := val.AuthzGrant(ctx, granter, grantee.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	sendAmount := 5
	nestedCmd := []string{
		noble.Config().Bin,
		"tx", "bank", "send", granter.FormattedAddress(), receiver.FormattedAddress(), fmt.Sprintf("%duusdc", sendAmount),
		"--from", granter.FormattedAddress(), "--generate-only",
		"--chain-id", noble.GetNode().Chain.Config().ChainID,
		"--node", noble.GetNode().Chain.GetRPCAddress(),
		"--home", noble.GetNode().HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"--output", "json",
		"--yes",
	}

	// ACTION: Execute an authz SEND using a TF token from a grantee who is blacklisted
	// EXPECTED: Request fails; Even though grantee is acting on behalf of the granter,
	// the granter still cannot execute `send` due to being blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, grantee)

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens: unauthorized", grantee.FormattedAddress()))

	bal, err := noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, bal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, grantee)

	// ACTION: Execute an authz SEND using a TF token from a granter who is blacklisted
	// EXPECTED: Request fails; Granter is blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, granter)

	preSendBal := bal

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens: unauthorized", granter.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, bal.IsZero())
	// bal should not change
	// require.Equal(t, preSendBal, bal)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, granter)

	// ACTION: Execute an authz SEND using a TF token to a receiver who is blacklisted
	// EXPECTED: Request fails; Granter is blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver)

	preSendBal = bal

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens: unauthorized", receiver.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	// bal should not change
	require.Equal(t, preSendBal, bal)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver)

	// ACTION: Execute an authz SEND using a TF token while the TF is paused
	// EXPECTED: Request fails; chain is paused
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	preSendBal = bal

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, "the chain is paused")

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	// bal should not change
	require.Equal(t, preSendBal, bal)
}

func TestFiatTFBankSend(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	alice := w[0]
	bob := w[1]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// ACTION: Send TF token while TF is paused
	// EXPECTED: Request fails; token not sent

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	amountToSend := ibc.WalletAmount{
		Address: bob.FormattedAddress(),
		Denom:   "uusdc",
		Amount:  math.OneInt(),
	}
	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, "the chain is paused")

	bobBal, err := noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Send TF token while FROM address is blacklisted
	// EXPECTED: Request fails; token not sent

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, alice)

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens", alice.FormattedAddress()))

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, alice)

	// ACTION: Send TF token while TO address is blacklisted
	// EXPECTED: Request fails; token not sent

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, bob)

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens", bob.FormattedAddress()))

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, bob)

	// ACTION: Successfully send TF token
	// EXPECTED: Success

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.NoError(t, err, "error sending funds")

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, amountToSend.Amount, bobBal)
}
