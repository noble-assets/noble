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

package upgrade

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	ismkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	authoritytypes "github.com/noble-assets/authority/types"
	novaismtypes "github.com/noble-assets/nova/types/ism"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	consensusKeeper consensuskeeper.Keeper,
	ismKeeper ismkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		err = EnableVoteExtensions(ctx, logger, plan, consensusKeeper)
		if err != nil {
			return vm, err
		}

		err = ConfigureHyperlaneRoutingISM(ctx, logger, ismKeeper)
		if err != nil {
			return vm, err
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

// EnableVoteExtensions updates the x/consensus module parameters to enable
// vote extensions 5 blocks (~7.5 seconds) after the upgrade height.
func EnableVoteExtensions(ctx context.Context, logger log.Logger, upgradePlan upgradetypes.Plan, consensusKeeper consensuskeeper.Keeper) error {
	params, err := consensusKeeper.ParamsStore.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get consensus params from state")
	}

	height := upgradePlan.Height + 5
	params.Abci = &cmttypes.ABCIParams{
		VoteExtensionsEnableHeight: height,
	}

	err = consensusKeeper.ParamsStore.Set(ctx, params)
	if err != nil {
		return errors.Wrap(err, "unable to set consensus params to state")
	}

	logger.Info(fmt.Sprintf("vote extensions will be enabled at height %d", height))

	return nil
}

// ConfigureHyperlaneRoutingISM updates the default Hyperlane Routing ISM to
// utilize the x/nova ISM for the Noble Applayer domain.
func ConfigureHyperlaneRoutingISM(ctx context.Context, logger log.Logger, ismKeeper ismkeeper.Keeper) error {
	chainID := sdk.UnwrapSDKContext(ctx).ChainID()

	var applayerDomain int
	switch chainID {
	case DevnetChainID:
		applayerDomain = ApplayerDevnetChainID
	case TestnetChainID:
		applayerDomain = ApplayerTestnetChainID
	case MainnetChainID:
		applayerDomain = ApplayerMainnetChainID
	default:
		return fmt.Errorf("cannot configure hyperlane routing ism on %s chain", chainID)
	}

	err := ismKeeper.SetRoutingIsmDomain(ctx, &ismtypes.MsgSetRoutingIsmDomain{
		IsmId: DefaultISM,
		Route: ismtypes.Route{
			Ism:    novaismtypes.ExpectedId,
			Domain: uint32(applayerDomain),
		},
		Owner: authoritytypes.ModuleAddress.String(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to set hyperlane routing ism")
	}

	logger.Info("configured hyperlane routing ism with applayer ism", "domain", applayerDomain)

	return nil
}
