// Copyright 2024 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"context"
	"testing"
	"time"

	sdkupgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authoritytypes "github.com/noble-assets/authority/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	haltHeightDelta    = 10 // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = 10
)

type ChainUpgrade struct {
	Image       ibc.DockerImage
	UpgradeName string // if upgradeName is empty, assumes patch/rolling update
	Emergency   bool
	PreUpgrade  func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet)
	PostUpgrade func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet)
}

func TestChainUpgrade(
	t *testing.T,
	genesisVersion string,
	upgrades []ChainUpgrade,
) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	genesisImage := []ibc.DockerImage{
		{
			Repository: ghcrRepo,
			Version:    genesisVersion,
			UIDGID:     containerUidGid,
		},
	}

	nw, client := NobleSpinUp(t, ctx, genesisImage, false)
	noble := nw.Chain
	authority := nw.Authority

	logger := zaptest.NewLogger(t)

	for _, upgrade := range upgrades {
		if upgrade.PreUpgrade != nil {
			upgrade.PreUpgrade(t, ctx, noble, authority)
		}

		if upgrade.UpgradeName == "" {
			// patch/rolling upgrade
			if upgrade.Emergency {
				err := noble.StopAllNodes(ctx)
				require.NoError(t, err, "could not stop nodes for emergency upgrade")

				noble.UpgradeVersion(ctx, client, upgrade.Image.Repository, upgrade.Image.Version)

				err = noble.StartAllNodes(ctx)
				require.NoError(t, err, "could not start nodes for emergency upgrade")

				timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
				defer timeoutCtxCancel()

				err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
				require.NoError(t, err, "chain did not produce blocks after emergency upgrade")
			} else {
				// stage new version
				for _, n := range noble.Nodes() {
					n.Image = upgrade.Image
				}
				noble.UpgradeVersion(ctx, client, upgrade.Image.Repository, upgrade.Image.Version)

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

				err := testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
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

			upgradePlan, err := tx.SetMsgs([]sdk.Msg{
				&sdkupgradetypes.MsgSoftwareUpgrade{
					Authority: authoritytypes.ModuleAddress.String(),
					Plan: sdkupgradetypes.Plan{
						Name:   upgrade.UpgradeName,
						Height: haltHeight,
						Info:   upgrade.UpgradeName + " chain upgrade",
					},
				},
			})
			require.NoError(t, err)

			_, err = cosmos.BroadcastTx(
				ctx,
				broadcaster,
				authority,
				&authoritytypes.MsgExecute{
					Signer:   authority.FormattedAddress(),
					Messages: upgradePlan,
				},
			)
			require.NoError(t, err, "error submitting software upgrade tx")

			stdout, stderr, err := noble.Validators[0].ExecQuery(ctx, "upgrade", "plan")
			require.NoError(t, err, "error submitting software upgrade tx")

			logger.Debug("Upgrade", zap.String("plan_stdout", string(stdout)), zap.String("plan_stderr", string(stderr)))

			timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*10)
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
				n.Image = upgrade.Image
			}
			noble.UpgradeVersion(ctx, client, upgrade.Image.Repository, upgrade.Image.Version)

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

		if upgrade.PostUpgrade != nil {
			upgrade.PostUpgrade(t, ctx, noble, authority)
		}
	}
}
