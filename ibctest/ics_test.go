package ibctest_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ibctest "github.com/strangelove-ventures/ibctest/v3"
	"github.com/strangelove-ventures/ibctest/v3/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v3/ibc"
	"github.com/strangelove-ventures/ibctest/v3/relayer"
	"github.com/strangelove-ventures/ibctest/v3/relayer/rly"
	"github.com/strangelove-ventures/ibctest/v3/testreporter"
	"github.com/strangelove-ventures/ibctest/v3/testutil"
	integration "github.com/strangelove-ventures/noble/ibctest"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// This tests Cosmos Interchain Security, spinning up a provider and a single consumer chain.
func TestICS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	repo, version := integration.GetDockerImageInfo()

	ownerAddress_ := "noble1nye3jsmqf3v2wag09sfwzm30fl0lpfmjhczchm"
	var noble *cosmos.CosmosChain
	var roles NobleRoles
	var err error

	// Chain Factory
	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*ibctest.ChainSpec{
		{Name: "gaia", Version: "v9.0.0-rc1", ChainConfig: ibc.ChainConfig{GasAdjustment: 1.5}},
		{Name: "noble", ChainConfig: ibc.ChainConfig{
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
			PreGenesis: func(cc ibc.ChainConfig) error {
				val := noble.Validators[0]
				roles, err = noblePreGenesis(ctx, val)
				if err != nil {
					return err
				}
				return nil
			},
			// ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			// 	return modifyGenesisNobleAll(b, roles.Authority.Address, roles.Owner.Address, roles.MasterMinter.Address, roles.Blacklister.Address, roles.Pauser.Address)
			// },
			ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
				return modifyGenesisNobleAll(b, ownerAddress_, ownerAddress_, ownerAddress_, ownerAddress_, ownerAddress_)
			},
		}},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	provider, noble := chains[0], chains[1].(*cosmos.CosmosChain)

	// Relayer Factory
	client, network := ibctest.DockerSetup(t)
	r := ibctest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "andrew-paths_update", rly.RlyDefaultUidGid),
	).Build(t, client, network)

	// Prep Interchain
	const ibcPath = "ics-path"
	ic := ibctest.NewInterchain().
		AddChain(provider).
		AddChain(noble).
		AddRelayer(r, "relayer").
		AddProviderConsumerLink(ibctest.ProviderConsumerLink{
			Provider: provider,
			Consumer: noble,
			Relayer:  r,
			Path:     ibcPath,
		})

	// Log location
	f, err := ibctest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	err = ic.Build(ctx, eRep, ibctest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	})
	require.NoError(t, err, "failed to build interchain")

	err = testutil.WaitForBlocks(ctx, 10, provider, noble)
	require.NoError(t, err, "failed to wait for blocks")

	nobleValidator := noble.Validators[0]

	_, err = nobleValidator.ExecTx(ctx, masterMinterKeyName,
		"tokenfactory", "configure-minter-controller", roles.MinterController.Address, roles.Minter.Address,
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, minterControllerKeyName,
		"tokenfactory", "configure-minter", roles.Minter.Address, "1000uusdc",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, roles.User.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "failed to mint uusdc to user")

	_, err = nobleValidator.ExecTx(ctx, blacklisterKeyName,
		"tokenfactory", "blacklist", roles.User.Address,
	)
	require.NoError(t, err, "failed to blacklist user address")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User2.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user2 tx")

	err = nobleValidator.SendFunds(ctx, user2KeyName, ibc.WalletAmount{
		Address: roles.User.Address,
		Denom:   "uusdc",
		Amount:  50,
	})
	require.Error(t, err, "The tx to a blacklisted user should not have been successful")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "uusdc")
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
	require.Equal(t, int64(10_100), userBalance, "user balance should have incremented")

	_, err = nobleValidator.ExecTx(ctx, blacklisterKeyName,
		"tokenfactory", "unblacklist", roles.User.Address,
	)
	require.NoError(t, err, "failed to unblacklist user address")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(200), userBalance, "user balance should have increased now that they are no longer blacklisted")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.Minter.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	minterBalance, err := noble.GetBalance(ctx, roles.Minter.Address, "uusdc")
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, int64(100), minterBalance, "minter balance should have increased")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName, "tokenfactory", "burn", "10uusdc")
	require.NoError(t, err, "failed to execute burn tx")

	minterBalance, err = noble.GetBalance(ctx, roles.Minter.Address, "uusdc")
	require.NoError(t, err, "failed to get minter balance")
	require.Equal(t, int64(90), minterBalance, "minter balance should have decreased because tokens were burned")

	_, err = nobleValidator.ExecTx(ctx, pauserKeyName,
		"tokenfactory", "pause",
	)
	require.NoError(t, err, "failed to pause mints")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", roles.User.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, userKeyName,
		"bank", "send", roles.User.Address, roles.Alice.Address, "100uusdc",
	)
	require.Error(t, err, "transaction was successful while chain was paused")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err := noble.GetBalance(ctx, roles.Alice.Address, "uusdc")
	require.NoError(t, err, "failed to get alice balance")

	require.Equal(t, int64(0), aliceBalance, "alice balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName, "tokenfactory", "burn", "10uusdc")
	require.NoError(t, err, "failed to execute burn tx")
	require.Equal(t, int64(90), minterBalance, "this burn should not have been successful because the chain is paused")

	_, err = nobleValidator.ExecTx(ctx, masterMinterKeyName,
		"tokenfactory", "configure-minter-controller", roles.MinterController.Address, roles.User.Address,
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, _, err = nobleValidator.ExecQuery(ctx, "tokenfactory", "show-minters", roles.User.Address)
	require.Error(t, err, "'user' should not have been able to become a minter while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, minterControllerKeyName, "tokenfactory", "remove-minter", roles.Minter.Address)
	require.NoError(t, err, "failed to send remove minter tx")

	_, _, err = nobleValidator.ExecQuery(ctx, "tokenfactory", "show-minters", roles.Minter.Address)
	require.Error(t, err, "minter should have been removed, even while chain is puased")

	_, err = nobleValidator.ExecTx(ctx, pauserKeyName,
		"tokenfactory", "unpause",
	)
	require.NoError(t, err, "failed to unpause mints")

	_, err = nobleValidator.ExecTx(ctx, userKeyName,
		"bank", "send", roles.User.Address, roles.Alice.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to send tx bank from user to alice")

	userBalance, err = noble.GetBalance(ctx, roles.User.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(100), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err = noble.GetBalance(ctx, roles.Alice.Address, "uusdc")
	require.NoError(t, err, "failed to get alice balance")
	require.Equal(t, int64(100), aliceBalance, "alice balance should not have increased while chain is paused")

}
