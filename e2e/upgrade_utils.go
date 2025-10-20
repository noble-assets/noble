// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
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
	"fmt"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdkupgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/docker/docker/client"
	authoritytypes "github.com/noble-assets/authority/types"
	"github.com/strangelove-ventures/interchaintest/v8"
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
	PreUpgrade  func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *ICATestSuite)
	PostUpgrade func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *ICATestSuite)
}

func GhcrImage(version string) ibc.DockerImage {
	return ibc.DockerImage{
		Repository: ghcrRepo,
		Version:    version,
		UIDGID:     containerUidGid,
	}
}

func TestChainUpgrade(
	t *testing.T,
	genesisVersion string,
	upgrades []ChainUpgrade,
	testICA bool,
) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	genesisImage := []ibc.DockerImage{GhcrImage(genesisVersion)}

	var (
		nw     NobleWrapper
		client *client.Client
		icaTs  *ICATestSuite
		err    error
	)

	switch {
	case testICA:
		nw, client, icaTs, err = setupICATestSuite(t, ctx, genesisImage)
		require.NoError(t, err)
	default:
		nw, client = NobleSpinUp(t, ctx, genesisImage, false)
	}

	noble := nw.Chain
	authority := nw.Authority

	logger := zaptest.NewLogger(t)

	for _, upgrade := range upgrades {
		if upgrade.PreUpgrade != nil {
			upgrade.PreUpgrade(t, ctx, noble, authority, icaTs)
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

			// Upgrades prior to and including the helium upgrade require the use of the old paramauthority module.
			switch {
			case upgrade.UpgradeName == "helium":
				err := submitPreV8UpgradeTx(t, ctx, noble, upgrade, authority, haltHeight)
				require.NoError(t, err, "error submitting software upgrade tx")
			default:
				err := submitPostV8UpgradeTx(t, ctx, noble, upgrade, authority, haltHeight)
				require.NoError(t, err, "error submitting software upgrade tx")
			}

			stdout, stderr, err := noble.Validators[0].ExecQuery(ctx, "upgrade", "plan")
			require.NoError(t, err, "error submitting software upgrade tx")

			logger.Debug("Upgrade", zap.String("plan_stdout", string(stdout)), zap.String("plan_stderr", string(stderr)))

			timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*20)
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

			out, outE, err := noble.GetFullNode().ExecQuery(ctx, "orbiter", "forwarder", "paused-protocols")
			require.NoError(t, err, "failed unexpectedly")

			require.Contains(t, string(out), "Hyperlane")
			require.Empty(t, string(outE))
		}

		if upgrade.PostUpgrade != nil {
			upgrade.PostUpgrade(t, ctx, noble, authority, icaTs)
		}
	}
}

// setupICATestSuite attempts to spin up Noble and a counterparty chain, initialize an IBC path between the two chains,
// and setup any other necessary values that are required for testing interchain accounts.
func setupICATestSuite(
	t *testing.T,
	ctx context.Context,
	genesisImage []ibc.DockerImage,
) (NobleWrapper, *client.Client, *ICATestSuite, error) {
	t.Helper()

	nw, ibcSimd, _, r, ibcPathName, _, eRep, client, _ := NobleSpinUpIBC(t, ctx, genesisImage, false)

	err := r.StartRelayer(ctx, eRep, ibcPathName)
	if err != nil {
		return NobleWrapper{}, nil, nil, fmt.Errorf("error starting relayer: %w", err)
	}

	t.Cleanup(func() {
		_ = r.StopRelayer(ctx, eRep)
	})

	connections, err := r.GetConnections(ctx, eRep, ibcSimd.Config().ChainID)
	if err != nil {
		return NobleWrapper{}, nil, nil, fmt.Errorf("error querying ibc connections on %s: %w", ibcSimd.Config().ChainID, err)
	}
	if len(connections) == 0 {
		return NobleWrapper{}, nil, nil, fmt.Errorf("ibc connection not found on chain: %s", ibcSimd.Config().ChainID)
	}

	controllerConnectionID := connections[0].ID

	connections, err = r.GetConnections(ctx, eRep, nw.Chain.Config().ChainID)
	if err != nil {
		return NobleWrapper{}, nil, nil, fmt.Errorf("error querying ibc connections on %s: %w", nw.Chain.Config().ChainID, err)
	}
	if len(connections) == 0 {
		return NobleWrapper{}, nil, nil, fmt.Errorf("ibc connection not found on chain: %s", nw.Chain.Config().ChainID)
	}

	hostConnectionID := connections[0].ID

	initBal := math.NewInt(int64(10_000_000))

	users := interchaintest.GetAndFundTestUsers(t, ctx, "user", initBal, ibcSimd)
	ownerAddress := users[0].FormattedAddress()

	icaTs := &ICATestSuite{
		Host:                   nw.Chain,
		Controller:             ibcSimd,
		Relayer:                r,
		Rep:                    eRep,
		OwnerAddress:           ownerAddress,
		InitBal:                initBal,
		HostConnectionID:       hostConnectionID,
		ControllerConnectionID: controllerConnectionID,
		Encoding:               "proto3",
	}

	return nw, client, icaTs, nil
}

// submitPreV8UpgradeTx attempts to submit the software upgrade tx for pre Noble v8.0.0 releases.
// The software upgrade tx logic is compatible with the old paramauthority module.
func submitPreV8UpgradeTx(
	t *testing.T,
	ctx context.Context,
	noble *cosmos.CosmosChain,
	upgrade ChainUpgrade,
	authority ibc.Wallet,
	haltHeight int64,
) error {
	t.Helper()

	cmd := []string{
		"upgrade", "software-upgrade", upgrade.UpgradeName,
		"--upgrade-height", strconv.Itoa(int(haltHeight)),
		"--upgrade-info", "",
	}

	_, err := noble.Validators[0].ExecTx(ctx, authority.KeyName(), cmd...)
	if err != nil {
		return err
	}

	return nil
}

// submitPostV8UpgradeTx attempts to submit the software upgrade tx for Noble v8.0.0 and later releases.
// The software upgrade tx logic is compatible with the Noble Authority module.
func submitPostV8UpgradeTx(
	t *testing.T,
	ctx context.Context,
	noble *cosmos.CosmosChain,
	upgrade ChainUpgrade,
	authority ibc.Wallet,
	haltHeight int64,
) error {
	t.Helper()

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
	if err != nil {
		return err
	}

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		authority,
		&authoritytypes.MsgExecute{
			Signer:   authority.FormattedAddress(),
			Messages: upgradePlan,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
