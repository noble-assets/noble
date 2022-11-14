package cosmos_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/ibctest/v3"
	"github.com/strangelove-ventures/ibctest/v3/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v3/ibc"
	"github.com/strangelove-ventures/ibctest/v3/test"
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

	t.Parallel()

	ctx := context.Background()

	client, network := ibctest.DockerSetup(t)

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "token",
		Bech32Prefix:   "cosmos",
		GasPrices:      "0.0token",
		GasAdjustment:  1.1,
		TrustingPeriod: "504h",
		NoHostMount:    false,
		Images: []ibc.DockerImage{
			{
				Repository: "noble",
				Version:    "v0.45.10",
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

	kr := keyring.NewInMemory()

	masterMinter := ibctest.BuildWallet(kr, masterMinterKeyName, chainCfg)
	minter := ibctest.BuildWallet(kr, minterKeyName, chainCfg)

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
