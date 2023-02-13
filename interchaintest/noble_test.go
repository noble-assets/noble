package interchaintest_test

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	integration "github.com/strangelove-ventures/noble/interchaintest"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNobleChain(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	repo, version := integration.GetDockerImageInfo()

	var noble *cosmos.CosmosChain
	var roles NobleRoles

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "token",
		Bech32Prefix:   "noble",
		CoinType:       "118",
		GasPrices:      "0.0token",
		GasAdjustment:  1.1,
		TrustingPeriod: "504h",
		NoHostMount:    false,
		Images: []ibc.DockerImage{
			{
				Repository: repo,
				Version:    version,
				UidGid:     "1025:1025",
			},
		},
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) (err error) {
			val := noble.Validators[0]
			roles, err = noblePreGenesis(ctx, val)
			return err
		},
		ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			return modifyGenesisNobleOwner(b, roles.Owner.Address)
		},
	}

	nv := 2
	nf := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			ChainConfig:   chainCfg,
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	noble = chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	nobleValidator := noble.Validators[0]

	_, err = nobleValidator.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-master-minter", roles.MasterMinter.Address,
	)
	require.NoError(t, err, "failed to execute update master minter tx")

	_, err = nobleValidator.ExecTx(ctx, masterMinterKeyName,
		"tokenfactory", "configure-minter-controller", roles.MinterController.Address, roles.Minter.Address,
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, minterControllerKeyName,
		"tokenfactory", "configure-minter", roles.Minter.Address, "1000urupee",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100urupee",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "failed to mint urupee to user")

	_, err = nobleValidator.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-blacklister", roles.Blacklister.Address,
	)
	require.NoError(t, err, "failed to set blacklister")

	_, err = nobleValidator.ExecTx(ctx, blacklisterKeyName,
		"tokenfactory", "blacklist", roles.User.Address,
	)
	require.NoError(t, err, "failed to blacklist user address")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100urupee",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User2.Address, "100urupee",
	)
	require.NoError(t, err, "failed to execute mint to user2 tx")

	err = nobleValidator.SendFunds(ctx, user2KeyName, ibc.WalletAmount{
		Address: roles.User.Address,
		Denom:   "urupee",
		Amount:  50,
	})
	require.Error(t, err, "The tx to a blacklisted user should not have been successful")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	err = nobleValidator.SendFunds(ctx, user2KeyName, ibc.WalletAmount{
		Address: roles.User.Address,
		Denom:   "token",
		Amount:  100,
	})
	require.NoError(t, err, "The tx should have been successfull as that is no the minting denom")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "token")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should have incremented")

	_, err = nobleValidator.ExecTx(ctx, blacklisterKeyName,
		"tokenfactory", "unblacklist", roles.User.Address,
	)
	require.NoError(t, err, "failed to unblacklist user address")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100urupee",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(200), userBalance, "user balance should have increased now that they are no longer blacklisted")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.Minter.Address, "100urupee",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	minterBalance, err := noble.GetBalance(ctx, roles.Minter.Address, "urupee")
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, int64(100), minterBalance, "minter balance should have increased")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName, "tokenfactory", "burn", "10urupee")
	require.NoError(t, err, "failed to execute burn tx")

	minterBalance, err = noble.GetBalance(ctx, roles.Minter.Address, "urupee")
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, int64(90), minterBalance, "minter balance should have decreased because tokens were burned")

	_, err = nobleValidator.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-pauser", roles.Pauser.Address,
	)
	require.NoError(t, err, "failed to update pauser")

	_, err = nobleValidator.ExecTx(ctx, pauserKeyName,
		"tokenfactory", "pause",
	)
	require.NoError(t, err, "failed to pause mints")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100urupee",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, userKeyName,
		"bank", "send", roles.User.Address, roles.Alice.Address, "100urupee",
	)
	require.Error(t, err, "transaction was successful while chain was paused")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err := noble.GetBalance(ctx, roles.Alice.Address, "urupee")
	require.NoError(t, err, "failed to get alice balance")

	require.Equal(t, int64(0), aliceBalance, "alice balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName, "tokenfactory", "burn", "10urupee")
	require.NoError(t, err, "failed to execute burn tx")
	require.Equal(t, int64(90), minterBalance, "this burn should not have been successful because the chain is paused")

	_, err = nobleValidator.ExecTx(ctx, masterMinterKeyName,
		"tokenfactory", "configure-minter-controller", roles.MinterController.Address, roles.User.Address,
	)
	require.NoError(t, err, "tx to configure minter controller should still succeed even though in paused state")

	_, _, err = nobleValidator.ExecQuery(ctx, "tokenfactory", "show-minters", roles.User.Address)
	require.Error(t, err, "'user' should not have been able to become a minter while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, minterControllerKeyName, "tokenfactory", "remove-minter", roles.Minter.Address)
	require.NoError(t, err, "minters should be able to be removed while in paused state")

	_, _, err = nobleValidator.ExecQuery(ctx, "tokenfactory", "show-minters", roles.Minter.Address)
	require.Error(t, err, "minter should not have been added so remove should be a no-op and this should fail")

	_, err = nobleValidator.ExecTx(ctx, pauserKeyName,
		"tokenfactory", "unpause",
	)
	require.NoError(t, err, "failed to unpause mints")

	_, err = nobleValidator.ExecTx(ctx, userKeyName,
		"bank", "send", roles.User.Address, roles.Alice.Address, "100urupee",
	)
	require.NoError(t, err, "failed to send tx bank from user to alice")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "urupee")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err = noble.GetBalance(ctx, roles.Alice.Address, "urupee")
	require.NoError(t, err, "failed to get alice balance")
	require.Equal(t, int64(100), aliceBalance, "alice balance should not have increased while chain is paused")

}
