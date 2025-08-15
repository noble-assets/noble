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
	cmttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	consensusKeeper consensuskeeper.Keeper,
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
