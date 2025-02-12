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
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dollarkeeper "dollar.noble.xyz/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	dollarKeeper *dollarkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		chainID := sdk.UnwrapSDKContext(ctx).ChainID()
		if chainID != TestnetChainID {
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, chainID)
		}

		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		return vm, MigrateDollarPortalState(ctx, dollarKeeper)
	}
}

// MigrateDollarPortalState migrates the state of the Noble Dollar Portal.
func MigrateDollarPortalState(ctx context.Context, dollarKeeper *dollarkeeper.Keeper) (err error) {
	err = dollarKeeper.PortalOwner.Set(ctx, "noble1mx48c5tv6ss9k7793n3a7sv48nfjllhxkd6tq3")
	if err != nil {
		return errors.Wrap(err, "unable to migrate dollar portal owner")
	}

	err = dollarKeeper.PortalNonce.Set(ctx, 1)
	if err != nil {
		return errors.Wrap(err, "unable to migrate dollar portal nonce")
	}

	return nil
}
