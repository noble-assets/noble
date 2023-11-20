package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkupgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/strangelove-ventures/noble/cmd"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	ghcrRepo        = "ghcr.io/strangelove-ventures/noble"
	containerUidGid = "1025:1025"

	haltHeightDelta    = uint64(10) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = uint64(10)
)

type ParamsQueryResponse struct {
	Subspace string `json:"subspace"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type chainUpgrade struct {
	image       ibc.DockerImage
	upgradeName string // if upgradeName is empty, assumes patch/rolling update
	emergency   bool
	preUpgrade  func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, paramAuthority ibc.Wallet)
	postUpgrade func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, paramAuthority ibc.Wallet)
}

func ghcrImage(version string) ibc.DockerImage {
	return ibc.DockerImage{
		Repository: ghcrRepo,
		Version:    version,
		UidGid:     containerUidGid,
	}
}

func testNobleChainUpgrade(
	t *testing.T,
	chainID string,
	genesisVersionImage ibc.DockerImage,
	genesisTokenFactoryDenomMetadata DenomMetadata,
	numberOfValidators int,
	numberOfFullNodes int,
	upgrades []chainUpgrade,
) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var gw genesisWrapper

	cs := nobleChainSpec(ctx, &gw, chainID, numberOfValidators, numberOfFullNodes, false, false, false, false)

	cs.ChainConfig.PreGenesis = func(cc ibc.ChainConfig) error {
		val := gw.chain.Validators[0]
		var err error
		gw.tfRoles, err = createTokenfactoryRoles(ctx, genesisTokenFactoryDenomMetadata, val, true)
		if err != nil {
			return err
		}
		gw.paramAuthority, err = createParamAuthAtGenesis(ctx, val)
		return err
	}

	cs.ChainConfig.ModifyGenesis = func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(b, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}
		if err := modifyGenesisTokenfactory(g, "tokenfactory", genesisTokenFactoryDenomMetadata, gw.tfRoles, true); err != nil {
			return nil, err
		}
		if err := modifyGenesisParamAuthority(g, gw.paramAuthority.FormattedAddress()); err != nil {
			return nil, err
		}
		if err := modifyGenesisDowntimeWindow(g); err != nil {
			return nil, err
		}
		out, err := json.Marshal(&g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}

	cs.ChainConfig.Images = []ibc.DockerImage{genesisVersionImage}

	logger := zaptest.NewLogger(t)

	cf := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{cs})
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
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	chainCfg := noble.Config()

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	for _, upgrade := range upgrades {
		if upgrade.preUpgrade != nil {
			upgrade.preUpgrade(t, ctx, noble, gw.paramAuthority)
		}

		if upgrade.upgradeName == "" {
			// patch/rolling upgrade
			if upgrade.emergency {
				err = noble.StopAllNodes(ctx)
				require.NoError(t, err, "could not stop nodes for emergency upgrade")

				noble.UpgradeVersion(ctx, client, upgrade.image.Repository, upgrade.image.Version)

				err = noble.StartAllNodes(ctx)
				require.NoError(t, err, "could not start nodes for emergency upgrade")

				timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
				defer timeoutCtxCancel()

				err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
				require.NoError(t, err, "chain did not produce blocks after emergency upgrade")
			} else {
				// stage new version
				for _, n := range noble.Nodes() {
					n.Image = upgrade.image
				}
				noble.UpgradeVersion(ctx, client, upgrade.image.Repository, upgrade.image.Version)

				// do rolling update on half the vals
				for i, n := range noble.Validators {
					if i%2 == 0 {
						continue
					}
					// shutdown
					require.NoError(t, n.StopContainer(ctx))
					require.NoError(t, n.RemoveContainer(ctx))

					// startup
					require.NoError(t, n.CreateNodeContainer(ctx))
					require.NoError(t, n.StartContainer(ctx))

					timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
					defer timeoutCtxCancel()

					require.NoError(t, testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble))
				}

				// blocks should still be produced after rolling update
				timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
				defer timeoutCtxCancel()

				err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
				require.NoError(t, err, "chain did not produce blocks after upgrade")

				// stop all nodes to bring rest of vals up to date
				err = noble.StopAllNodes(ctx)
				require.NoError(t, err, "error stopping node(s)")

				err = noble.StartAllNodes(ctx)
				require.NoError(t, err, "error starting upgraded node(s)")

				timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
				defer timeoutCtxCancel()

				err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
				require.NoError(t, err, "chain did not produce blocks after upgrade")
			}
		} else {
			// halt upgrade
			height, err := noble.Height(ctx)
			require.NoError(t, err, "error fetching height before submit upgrade proposal")

			haltHeight := height + haltHeightDelta

			broadcaster := cosmos.NewBroadcaster(t, noble)

			upgradePlan := sdkupgradetypes.Plan{
				Name:   upgrade.upgradeName,
				Height: int64(haltHeight),
				Info:   upgrade.upgradeName + " chain upgrade",
			}

			wallet := cosmos.NewWallet(
				gw.paramAuthority.KeyName(),
				gw.paramAuthority.Address(),
				gw.paramAuthority.Mnemonic(),
				chainCfg,
			)

			_, err = cosmos.BroadcastTx(
				ctx,
				broadcaster,
				wallet,
				&upgradetypes.MsgSoftwareUpgrade{
					Authority: gw.paramAuthority.FormattedAddress(),
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

			// upgrade all nodes
			for _, n := range noble.Nodes() {
				n.Image = upgrade.image
			}
			noble.UpgradeVersion(ctx, client, upgrade.image.Repository, upgrade.image.Version)

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

		if upgrade.postUpgrade != nil {
			upgrade.postUpgrade(t, ctx, noble, gw.paramAuthority)
		}
	}
}
