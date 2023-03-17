package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkupgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/interchaintest/v3/testutil"
	"github.com/strangelove-ventures/noble/cmd"
	integration "github.com/strangelove-ventures/noble/interchaintest"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	haltHeightDelta    = uint64(10) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = uint64(10)
)

func TestNobleChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	_, version := integration.GetDockerImageInfo()

	var noble *cosmos.CosmosChain
	var roles NobleRoles
	var paramauthorityWallet Authority

	const (
		upgradeName = "neon"
	)

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
				Repository: "ghcr.io/strangelove-ventures/noble",
				Version:    "v0.3.0",
				UidGid:     "1025:1025",
			},
		},
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) error {
			val := noble.Validators[0]
			err := createTokenfactoryRoles(ctx, &roles, DenomMetadata_rupee, val, true)
			if err != nil {
				return err
			}
			paramauthorityWallet, err = createParamAuthAtGenesis(ctx, val)
			return err
		},
		ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			g := make(map[string]interface{})
			if err := json.Unmarshal(b, &g); err != nil {
				return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
			}
			err := modifyGenesisTokenfactory(g, "tokenfactory", DenomMetadata_rupee, &roles, true)
			if err != nil {
				return nil, err
			}
			err = modifyGenesisParamAuthority(g, paramauthorityWallet.Authority.Address)
			if err != nil {
				return nil, err
			}
			out, err := json.Marshal(&g)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
			}
			return out, nil
		},
	}

	nv := 2
	nf := 0

	logger := zaptest.NewLogger(t)

	cf := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
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

	height, err := noble.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	broadcaster := cosmos.NewBroadcaster(t, noble)

	upgradePlan := sdkupgradetypes.Plan{
		Name:   upgradeName,
		Height: int64(haltHeight),
		Info:   upgradeName + " chain upgrade",
	}

	decoded := sdk.MustAccAddressFromBech32(paramauthorityWallet.Authority.Address)
	wallet := &ibc.Wallet{
		Address:  string(decoded),
		Mnemonic: paramauthorityWallet.Authority.Mnemonic,
		KeyName:  paramauthorityWallet.Authority.KeyName,
		CoinType: paramauthorityWallet.Authority.CoinType,
	}

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		wallet,
		&upgradetypes.MsgSoftwareUpgrade{
			Authority: paramauthorityWallet.Authority.Address,
			Plan:      upgradePlan,
		},
	)
	require.NoError(t, err, "error submitting software upgrade tx")

	stdout, stderr, err := noble.Validators[0].ExecQuery(ctx, "upgrade", "plan")
	require.NoError(t, err, "error submitting software upgrade tx")

	logger.Debug("Upgrade", zap.String("plan_stdout", string(stdout)), zap.String("plan_stderr", string(stderr)))

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, noble)

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")

	// bring down nodes to prepare for upgrade
	err = noble.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version and repo on all nodes
	// TODO: fix local testing
	noble.UpgradeVersion(ctx, client, version)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	err = noble.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")

}
