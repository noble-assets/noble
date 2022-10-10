package cosmos_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/ibctest/v6"
	"github.com/strangelove-ventures/ibctest/v6/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v6/ibc"
	"github.com/strangelove-ventures/ibctest/v6/test"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	ownerKeyName  = "owner"
	ownerMnemonic = "member genre increase friend salmon nest seven custom improve cluster inform axis pact velvet hurt risk point worth excite fiscal omit romance grid evoke"

	masterMinterKeyName = "masterminter"
	minterKeyName       = "minter"
)

func NobleEncoding() *simappparams.EncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types

	return &cfg
}

func TestNobleChain(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	configFileOverrides := make(map[string]any)
	appTomlOverrides := make(ibc.ChainUtilToml)

	// state sync snapshots every stateSyncSnapshotInterval blocks.
	stateSync := make(ibc.ChainUtilToml)
	stateSync["snapshot-interval"] = 10
	appTomlOverrides["state-sync"] = stateSync

	// state sync snapshot interval must be a multiple of pruning keep every interval.
	appTomlOverrides["pruning"] = "nothing"

	configFileOverrides["config/app.toml"] = appTomlOverrides

	t.Parallel()

	ctx := context.Background()

	client, network := ibctest.DockerSetup(t)

	chainCfg := ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "noble",
		ChainID:             "noble-1",
		Bin:                 "nobled",
		Denom:               "token",
		Bech32Prefix:        "cosmos",
		GasPrices:           "0.0token",
		ConfigFileOverrides: configFileOverrides,
		GasAdjustment:       1.1,
		TrustingPeriod:      "504h",
		NoHostMount:         false,
		Images: []ibc.DockerImage{
			{
				Repository: "noble",
				Version:    "latest",
				UidGid:     "0:0",
			},
		},
		EncodingConfig: NobleEncoding(),
	}

	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*ibctest.ChainSpec{
		{
			Name:        "noble",
			ChainConfig: chainCfg,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	noble := chains[0].(*cosmos.CosmosChain)

	err = noble.Initialize(ctx, t.Name(), client, network)
	require.NoError(t, err, "failed to initialize noble chain")

	err = noble.RecoverKey(ctx, ownerKeyName, ownerMnemonic)
	require.NoError(t, err, "failed to recover owner key")

	ownerAddressBz, err := noble.GetAddress(ctx, ownerKeyName)
	require.NoError(t, err, "failed to get address for owner key")

	ownerAddress := types.MustBech32ifyAddressBytes(chainCfg.Bech32Prefix, ownerAddressBz)

	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	kr := keyring.NewInMemory(cdc)

	masterMinter := ibctest.BuildWallet(kr, masterMinterKeyName, chainCfg)
	minter := ibctest.BuildWallet(kr, masterMinterKeyName, chainCfg)

	ic := ibctest.NewInterchain().
		AddChain(noble,
			ibc.WalletAmount{
				Address: ownerAddress,
				Denom:   chainCfg.Denom,
				Amount:  100_000_000,
			},
			ibc.WalletAmount{
				Address: masterMinter.Address,
				Denom:   chainCfg.Denom,
				Amount:  100_000_000,
			},
			ibc.WalletAmount{
				Address: minter.Address,
				Denom:   chainCfg.Denom,
				Amount:  100_000_000,
			},
		)

	require.NoError(t, ic.Build(ctx, nil, ibctest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	require.NoError(t, err, "failed to start noble chain")

	nobleFullNode := noble.FullNodes[0]

	_, err = nobleFullNode.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-master-minter", masterMinter.Address,
	)
	require.NoError(t, err, "failed to execute update master minter tx")

	err = test.WaitForBlocks(ctx, 1, noble)
	require.NoError(t, err, "failed to wait for a block on noble chain")

}
