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
	"errors"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	authoritykeeper "github.com/noble-assets/authority/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		err = ClaimDistributionFunds(ctx, logger, addressCodec, authorityKeeper, bankKeeper)
		if err != nil {
			return vm, err
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

// ClaimDistributionFunds transfers all transaction fees accrued by Noble prior
// to the v8 Helium upgrade (November 2024) to the x/authority owner. The funds
// are currently stuck as the x/distribution module was removed and replaced by
// the x/authority module without a proper migration of funds.
func ClaimDistributionFunds(ctx context.Context, logger log.Logger, addressCodec address.Codec, authorityKeeper *authoritykeeper.Keeper, bankKeeper bankkeeper.Keeper) error {
	// NOTE: We hardcode the x/distribution module name to avoid an import.
	address := authtypes.NewModuleAddress("distribution")
	balance := bankKeeper.GetAllBalances(ctx, address)
	if balance.IsZero() {
		// We return early in the case that there are no claimable funds.
		return nil
	}

	authority, err := authorityKeeper.Owner.Get(ctx)
	if err != nil {
		return errors.New("unable to get underlying authority address from state")
	}
	authorityBz, err := addressCodec.StringToBytes(authority)
	if err != nil {
		return errors.New("unable to decode underlying authority address")
	}

	err = bankKeeper.SendCoins(ctx, address, authorityBz, balance)
	if err != nil {
		return errors.New("unable to transfer stuck distribution funds")
	}

	logger.Info("claimed stuck distribution module funds", "amount", balance.String())

	return nil
}
