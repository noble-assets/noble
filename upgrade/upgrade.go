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
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	dollarkeeper "dollar.noble.xyz/v2/keeper"
	dollartypes "dollar.noble.xyz/v2/types"
	authoritytypes "github.com/noble-assets/authority/types"
	swapkeeper "swap.noble.xyz/keeper"
	swaptypes "swap.noble.xyz/types"
	stableswaptypes "swap.noble.xyz/types/stableswap"
)

const RECIPIENT = "noble1c3chgrgr3xcktkpvezxz7g9kl7h64x8tdyd8ng"

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	addressCodec address.Codec,
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
			if err = claimSwapPoolsYield(ctx, logger, addressCodec, bankKeeper, dollarKeeper, swapKeeper); err != nil {
				return vm, err
			}

			if err = claimSwapPoolsProtocolFees(ctx, logger, swapKeeper); err != nil {
				return vm, err
			}

			if err = closeSwapPools(ctx, logger, swapKeeper); err != nil {
				return vm, err
			}
		}

		return vm, nil
	}
}

// claimSwapPoolsYield claims the $USDN yield accrued in all Noble Swap pools.
func claimSwapPoolsYield(
	ctx context.Context,
	logger log.Logger,
	addressCodec address.Codec,
	bankKeeper bankkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	swapKeeper *swapkeeper.Keeper,
) error {
	dollarServer := dollarkeeper.NewMsgServer(dollarKeeper)

	recipient, err := addressCodec.StringToBytes(RECIPIENT)
	if err != nil {
		return fmt.Errorf("unable to decode recipient address: %w", err)
	}

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

		amount := sdk.NewCoins(sdk.NewCoin(dollarKeeper.GetDenom(), yield))
		err = bankKeeper.SendCoins(ctx, address, recipient, amount)
		if err != nil {
			return fmt.Errorf("unable to transfer yield for pool %d", pool.Id)
		}

		logger.Info("claimed swap pool yield", "pool", pool.Id, "amount", amount.String())
	}

	return nil
}

// claimSwapPoolsProtocolFees claims the protocol fees accrued in all Noble Swap pools.
func claimSwapPoolsProtocolFees(
	ctx context.Context,
	logger log.Logger,
	swapKeeper *swapkeeper.Keeper,
) error {
	swapServer := swapkeeper.NewMsgServer(swapKeeper)

	amount := sdk.NewCoins()
	poolsRes, err := swapkeeper.NewQueryServer(swapKeeper).Pools(ctx, &swaptypes.QueryPools{})
	if err != nil {
		logger.Error("unable to get pools", "err", err)
	} else {
		for _, pool := range poolsRes.Pools {
			amount = amount.Add(pool.ProtocolFees...)
		}
	}

	_, err = swapServer.WithdrawProtocolFees(ctx, &swaptypes.MsgWithdrawProtocolFees{
		Signer: authoritytypes.ModuleAddress.String(),
		To:     RECIPIENT,
	})
	if err != nil {
		logger.Error("unable to withdraw protocol fees", "err", err)
	}

	logger.Info("claimed swap pool protocol fees", "amount", amount.String())

	return nil
}

// closeSwapPools force-unbonds every active liquidity provider and then permanently pauses the pool.
func closeSwapPools(
	ctx context.Context,
	logger log.Logger,
	swapKeeper *swapkeeper.Keeper,
) error {
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

		// Backdate each unbonding position's EndTime so it completes on the next BeginBlocker logic run (which we trigger next).
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
		logger.Info("removing swap pool liquidity", "pool", poolId, "address", userAddress, "shares", amount, "amount", userTotalAmount.String())
	}

	// Run the BeginBlocker logic to process the backdated unbondings.
	if err = swapKeeper.BeginBlocker(ctx); err != nil {
		return err
	}

	// Pause the pool for good.
	if err = swapKeeper.SetPaused(ctx, 0, true); err != nil {
		return err
	}

	return nil
}
