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
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	dollarkeeper "dollar.noble.xyz/v2/keeper"
	dollartypes "dollar.noble.xyz/v2/types"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	swapkeeper "swap.noble.xyz/keeper"
	swaptypes "swap.noble.xyz/types"
	stableswaptypes "swap.noble.xyz/types/stableswap"
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

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		if sdkCtx.ChainID() == MainnetChainID {
			if err = claimSwapPoolsYield(
				ctx,
				logger,
				addressCodec,
				authorityKeeper,
				bankKeeper,
				dollarKeeper,
				swapKeeper,
			); err != nil {
				return vm, err
			}

			if err = claimSwapPoolsProtocolFees(
				ctx,
				logger,
				swapKeeper,
				"noble1c3chgrgr3xcktkpvezxz7g9kl7h64x8tdyd8ng",
			); err != nil {
				return vm, err
			}

			if err = closeSwapPools(ctx, logger, swapKeeper); err != nil {
				return vm, err
			}
		}

		return vm, nil
	}
}

// claimSwapPoolsYield claims the $USDN yield accrued inside the Noble Swap
// pools and sends it to the authority address.
func claimSwapPoolsYield(
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

// claimSwapPoolsProtocolFees claims the protocol fees accrued inside the Noble Swap
// pools.
func claimSwapPoolsProtocolFees(
	ctx context.Context,
	logger log.Logger,
	swapKeeper *swapkeeper.Keeper,
	protocolFeesReceiver string,
) error {
	swapServer := swapkeeper.NewMsgServer(swapKeeper)

	_, err := swapServer.WithdrawProtocolFees(ctx, &swaptypes.MsgWithdrawProtocolFees{
		Signer: authtypes.NewModuleAddressOrBech32Address("authority").String(),
		To:     protocolFeesReceiver,
	})
	if err != nil {
		logger.Error(err.Error())
	}

	logger.Info("claimed swap protocol fees")

	return nil
}

// closeSwapPools force-unbonds every active liquidity provider and then permanently pauses the pools.
func closeSwapPools(ctx context.Context, logger log.Logger, swapKeeper *swapkeeper.Keeper) error {
	swapServer := swapkeeper.NewStableSwapMsgServer(swapKeeper)

	// Iterate over everyone with bonded shares, i.e. all the current liquidity providers.
	itr, err := swapKeeper.Stableswap.UsersTotalBondedShares.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	for ; itr.Valid(); itr.Next() {
		key, _ := itr.Key()
		poolId := key.K1()
		userAddress := key.K2()

		amount, amountErr := itr.Value()
		if amountErr != nil {
			return amountErr
		}

		// Skip users that have already removed all liquidity.
		if !amount.IsPositive() {
			continue
		}

		if _, err = swapServer.RemoveLiquidity(ctx, &stableswaptypes.MsgRemoveLiquidity{
			Signer:     userAddress,
			PoolId:     poolId,
			Percentage: math.LegacyNewDec(100), // 100%
		}); err != nil {
			return err
		}

		// Backdate each unbonding position's EndTime so it completes on the next BeginBlocker instead of days from now.
		unbondEndTime := sdk.UnwrapSDKContext(ctx).HeaderInfo().Time.Add(-24 * 3 * time.Hour)
		userTotalAmount := sdk.NewCoins()
		for _, position := range swapKeeper.Stableswap.GetUnbondingPositionsByProvider(ctx, userAddress) {
			if err = swapKeeper.Stableswap.RemoveUnbondingPosition(ctx, position.Timestamp, position.Address, position.PoolId); err != nil {
				return err
			}
			if err = swapKeeper.Stableswap.SetUnbondingPosition(ctx, unbondEndTime.Unix(), position.Address, position.PoolId, stableswaptypes.UnbondingPosition{
				Shares:  position.UnbondingPosition.Shares,
				Amount:  position.UnbondingPosition.Amount,
				EndTime: unbondEndTime,
			}); err != nil {
				return err
			}
			userTotalAmount = userTotalAmount.Add(position.UnbondingPosition.Amount...)
		}
		logger.Info("removing liquidy", "pool", poolId, "address", userAddress, "shares", amount.String(), "amount", userTotalAmount.String())
	}

	// Run the swap BeginBlocker to process the backdated unbondings.
	if err = swapKeeper.BeginBlocker(ctx); err != nil {
		return err
	}

	// Pause the pool for good; no one can interact with it after this.
	if err = swapKeeper.SetPaused(ctx, 0, true); err != nil {
		return err
	}

	return nil
}
