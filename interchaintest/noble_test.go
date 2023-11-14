package interchaintest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/noble/v5/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestNobleChain(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var gw genesisWrapper

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", 2, 1, false, false, true, true),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	t.Run("tokenfactory", func(t *testing.T) {
		t.Parallel()
		nobleTokenfactory_e2e(t, ctx, "tokenfactory", denomMetadataFrienzies.Base, noble, gw.tfRoles, gw.extraWallets)
	})

	t.Run("fiat-tokenfactory", func(t *testing.T) {
		t.Parallel()
		nobleTokenfactory_e2e(t, ctx, "fiat-tokenfactory", denomMetadataUsdc.Base, noble, gw.fiatTfRoles, gw.extraWallets)
	})
}

func nobleTokenfactory_e2e(t *testing.T, ctx context.Context, tokenfactoryModName, mintingDenom string, noble *cosmos.CosmosChain, roles NobleRoles, extraWallets ExtraWallets) {
	nobleValidator := noble.Validators[0]

	_, err := nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(), "-b", "block",
	)
	require.Error(t, err, "succeeded to execute update master minter tx by invalid owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-owner", roles.Owner2.FormattedAddress(), "-b", "block",
	)
	require.Error(t, err, "succeeded to execute update owner tx by invalid owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner.KeyName(),
		tokenfactoryModName, "update-owner", roles.Owner2.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute update owner tx")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(), "-b", "block",
	)
	require.Error(t, err, "succeeded to execute update master minter tx by pending owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "accept-owner", "-b", "block",
	)
	require.NoError(t, err, "failed to execute tx to accept ownership")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(), "-b", "block",
	)
	require.Error(t, err, "succeeded to execute update master minter tx by prior owner")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-master-minter", roles.MasterMinter.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute update master minter tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MasterMinter.KeyName(),
		tokenfactoryModName, "configure-minter-controller", roles.MinterController.FormattedAddress(), roles.Minter.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController.KeyName(),
		tokenfactoryModName, "configure-minter", roles.Minter.FormattedAddress(), "1000"+mintingDenom, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equalf(t, int64(100), userBalance, "failed to mint %s to user", mintingDenom)

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-blacklister", roles.Blacklister.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to set blacklister")

	_, err = nobleValidator.ExecTx(ctx, roles.Blacklister.KeyName(),
		tokenfactoryModName, "blacklist", extraWallets.User.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to blacklist user address")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.Error(t, err, "successfully executed mint to blacklisted user tx")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User2.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user2 tx")

	err = nobleValidator.SendFunds(ctx, extraWallets.User2.KeyName(), ibc.WalletAmount{
		Address: extraWallets.User.FormattedAddress(),
		Denom:   mintingDenom,
		Amount:  50,
	})
	require.Error(t, err, "The tx to a blacklisted user should not have been successful")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	err = nobleValidator.SendFunds(ctx, extraWallets.User2.KeyName(), ibc.WalletAmount{
		Address: extraWallets.User.FormattedAddress(),
		Denom:   "token",
		Amount:  100,
	})
	require.NoError(t, err, "The tx should have been successfull as that is no the minting denom")

	_, err = nobleValidator.ExecTx(ctx, roles.Blacklister.KeyName(),
		tokenfactoryModName, "unblacklist", extraWallets.User.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to unblacklist user address")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(200), userBalance, "user balance should have increased now that they are no longer blacklisted")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", roles.Minter.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	minterBalance, err := noble.GetBalance(ctx, roles.Minter.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, int64(100), minterBalance, "minter balance should have increased")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "burn", "10"+mintingDenom, "-b", "block",
	)
	require.NoError(t, err, "failed to execute burn tx")

	minterBalance, err = noble.GetBalance(ctx, roles.Minter.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, int64(90), minterBalance, "minter balance should have decreased because tokens were burned")

	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName(),
		tokenfactoryModName, "update-pauser", roles.Pauser.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to update pauser")

	// -- chain paused --

	_, err = nobleValidator.ExecTx(ctx, roles.Pauser.KeyName(),
		tokenfactoryModName, "pause", "-b", "block",
	)
	require.NoError(t, err, "failed to pause mints")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "mint", extraWallets.User.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.Error(t, err, "successfully executed mint to user tx while chain is paused")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, extraWallets.User.KeyName(),
		"bank", "send", extraWallets.User.FormattedAddress(), extraWallets.Alice.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.Error(t, err, "transaction was successful while chain is paused")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err := noble.GetBalance(ctx, extraWallets.Alice.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get alice balance")

	require.Equal(t, int64(0), aliceBalance, "alice balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, roles.Minter.KeyName(),
		tokenfactoryModName, "burn", "10"+mintingDenom, "-b", "block",
	)
	require.Error(t, err, "successfully executed burn tx while chain is paused")
	require.Equal(t, int64(90), minterBalance, "this burn should not have been successful because the chain is paused")

	_, err = nobleValidator.ExecTx(ctx, roles.MasterMinter.KeyName(),
		tokenfactoryModName, "configure-minter-controller", roles.MinterController2.FormattedAddress(), extraWallets.User.FormattedAddress(), "-b", "block")

	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController2.KeyName(),
		tokenfactoryModName, "configure-minter", extraWallets.User.FormattedAddress(), "1000"+mintingDenom, "-b", "block")
	require.NoError(t, err, "failed to execute configure minter tx")

	res, _, err := nobleValidator.ExecQuery(ctx, tokenfactoryModName, "show-minter-controller", roles.MinterController2.FormattedAddress(), "-o", "json")
	require.NoError(t, err, "failed to query minter controller")

	var minterControllerType types.QueryGetMinterControllerResponse
	json.Unmarshal(res, &minterControllerType)

	// minter controller should have been updated even while paused
	minterController2Address := roles.MinterController2.FormattedAddress()
	require.Equal(t, minterController2Address, minterControllerType.MinterController.Controller)

	// minter should have been updated even while paused
	userAddress := extraWallets.User.FormattedAddress()
	require.Equal(t, userAddress, minterControllerType.MinterController.Minter)

	_, err = nobleValidator.ExecTx(ctx, roles.MinterController2.KeyName(),
		tokenfactoryModName, "remove-minter", extraWallets.User.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "minters should be able to be removed while in paused state")

	_, err = nobleValidator.ExecTx(ctx, roles.Pauser.KeyName(),
		tokenfactoryModName, "unpause", "-b", "block",
	)
	require.NoError(t, err, "failed to unpause mints")

	// -- chain unpaused --

	_, err = nobleValidator.ExecTx(ctx, extraWallets.User.KeyName(),
		"bank", "send", extraWallets.User.FormattedAddress(), extraWallets.Alice.FormattedAddress(), "100"+mintingDenom, "-b", "block",
	)
	require.NoErrorf(t, err, "failed to send tx bank from user (%s) to alice (%s)", extraWallets.User.FormattedAddress(), extraWallets.Alice.FormattedAddress())

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err = noble.GetBalance(ctx, extraWallets.Alice.FormattedAddress(), mintingDenom)
	require.NoError(t, err, "failed to get alice balance")
	require.Equal(t, int64(100), aliceBalance, "alice balance should not have increased while chain is paused")
}
