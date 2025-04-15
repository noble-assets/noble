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
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dollarkeeper "dollar.noble.xyz/keeper"
	dollartypes "dollar.noble.xyz/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	swapkeeper "swap.noble.xyz/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	swapKeeper *swapkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		err = ClaimSwapPoolsYield(ctx, logger, addressCodec, authorityKeeper, bankKeeper, dollarKeeper, swapKeeper)
		if err != nil {
			return vm, err
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

// ClaimSwapPoolsYield claims the $USDN yield accrued inside the Noble Swap
// pools and sends it to the authority address.
func ClaimSwapPoolsYield(
	ctx context.Context,
	logger log.Logger,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	swapKeeper *swapkeeper.Keeper,
) error {
	authority, err := authorityKeeper.Owner.Get(ctx)
	if err != nil {
		return errors.New("unable to get underlying authority address from state")
	}
	authorityBz, err := addressCodec.StringToBytes(authority)
	if err != nil {
		return errors.New("unable to decode underlying authority address")
	}

	dollarServer := dollarkeeper.NewMsgServer(dollarKeeper)

	pools := swapKeeper.GetPools(ctx)
	for _, pool := range pools {
		yield, address, err := dollarKeeper.GetYield(ctx, pool.Address)
		if err != nil {
			return fmt.Errorf("unable to get yield for pool %d", pool.Id)
		}

		_, err = dollarServer.ClaimYield(ctx, &dollartypes.MsgClaimYield{Signer: pool.Address})
		if err != nil {
			return fmt.Errorf("unable to claim yield for pool %d", pool.Id)
		}

		err = bankKeeper.SendCoins(ctx, address, authorityBz, sdk.NewCoins(sdk.NewCoin(dollarKeeper.GetDenom(), yield)))
		if err != nil {
			return fmt.Errorf("unable to transfer yield for pool %d", pool.Id)
		}

		logger.Info("claimed swap pool yield", "pool", pool.Id, "yield", yield)
	}

	return nil
}
