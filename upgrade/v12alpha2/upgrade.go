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

package v12alpha2

import (
	"context"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	ftfkeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	ftftypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authoritytypes "github.com/noble-assets/authority/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	ftfKeeper *ftfkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		chainID := sdk.UnwrapSDKContext(ctx).ChainID()
		if chainID != DevnetChainID {
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, chainID)
		}

		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		// Because the paused store was not intialized in the devnet genesis,
		// it must be configured here for USDC transfers to be successful.
		ftfKeeper.SetPaused(ctx, ftftypes.Paused{Paused: false})

		// Because the owner store was not initialized in the devnet genesis,
		// it must be configured here for USDC issuance to be successful.
		ftfKeeper.SetOwner(ctx, ftftypes.Owner{Address: authoritytypes.ModuleAddress.String()})

		return vm, nil
	}
}
