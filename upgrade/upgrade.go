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

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	orbiterkeeper "github.com/noble-assets/orbiter/v2/keeper"
	dispatchercomp "github.com/noble-assets/orbiter/v2/keeper/component/dispatcher"
	orbitercore "github.com/noble-assets/orbiter/v2/types/core"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	accountKeeper *authkeeper.AccountKeeper,
	orbiterKeeper *orbiterkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		cachedCtx, writeCache := sdkCtx.CacheContext()
		err = updateOrbiterStats(cachedCtx, logger, orbiterKeeper)
		if err != nil {
			logger.Error("failed to updated Orbiter stats", "error", err)
		} else {
			writeCache()
		}

		cachedCtx, writeCache = sdkCtx.CacheContext()
		err = updateOrbiterModuleAccounts(cachedCtx, logger, *accountKeeper)
		if err != nil {
			logger.Error("failed to updated Orbiter module accounts", "error", err)
		} else {
			writeCache()
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

// The Orbiter module and dust collector accounts should be module accounts. If they received funds
// before the v11 upgrade, they remained stored as base account. This causes the query to the module
// account to fail. This handler migrates them to module accounts.
func updateOrbiterModuleAccounts(ctx sdk.Context, logger log.Logger, accountKeeper authkeeper.AccountKeeper) error {
	for _, name := range []string{orbitercore.ModuleName, orbitercore.DustCollectorName} {
		addr, perms := accountKeeper.GetModuleAddressAndPermissions(name)
		if addr == nil {
			return fmt.Errorf("failed to get module address and permissions for %s", name)
		}

		// Module account registration is lazy. When we run this function without real mainnet or
		// testnet data, the query will return `nil` because the modules are not registered yet.
		// We should skip this situation to perform e2e upgrade tests.
		acc := accountKeeper.GetAccount(ctx, addr)
		if acc == nil {
			continue
		}

		baseAcc, ok := (acc).(*authtypes.BaseAccount)
		if !ok {
			// We should skip the case in which the address is already associated with a module
			// address.
			_, ok := (acc).(*authtypes.ModuleAccount)
			if ok {
				logger.Info(fmt.Sprintf("skipped migration of %s, already a module account", name), "address", addr.String())
				continue
			}
			// If we are very unlucky...
			return fmt.Errorf("error creating the base account for %s: %T", name, acc)
		}

		macc := authtypes.NewModuleAccount(baseAcc, name, perms...)
		accountKeeper.SetModuleAccount(ctx, macc)
		logger.Info(fmt.Sprintf("migrated %s to a module account", name), "address", addr.String())
	}

	return nil
}

// updateOrbiterStats updates the statistics of the Orbiter module to improve readability of the
// denom and fix the used couterparty id. We have two cases to fix here:
//  1. Update entries that use the countertparty channel id OF Noble with the counterparty channel
//     id ON Noble (this error applies only for testnet since it is caused by the beta release)
//  2. Use the denom representation on Noble and not the IBC one. This means converting this
//     transfer/channel-4280/uusdc into uusdc
func updateOrbiterStats(ctx sdk.Context, logger log.Logger, orbiterKeeper *orbiterkeeper.Keeper) error {
	expectedDenom := "uusdc"
	channelsToCorrect := make(map[string]string)
	switch ctx.ChainID() {
	case MainnetChainID:
		// No-op for mainnet since channels are correct there.
	case TestnetChainID:
		// Counterparty channel to Noble -> Noble to counterparty channel.
		channelsToCorrect = map[string]string{
			"channel-4280": "channel-22",  // osmosis
			"channel-27":   "channel-639", // namada
			"channel-3":    "channel-333", // xion
			"channel-496":  "channel-43",  // neutron
		}
	default:
		return nil
	}

	dispatcher := orbiterKeeper.Dispatcher()
	if dispatcher == nil {
		return errors.New("received nil orbiter dispatcher component")
	}

	err := updateDispatchedAmounts(ctx, logger, expectedDenom, channelsToCorrect, dispatcher)
	if err != nil {
		return fmt.Errorf("failed to update dispatcher amounts: %w", err)
	}

	err = updateDispatchedCounts(ctx, logger, channelsToCorrect, dispatcher)
	if err != nil {
		return fmt.Errorf("failed to update dispatcher counts: %w", err)
	}

	return nil
}

func updateDispatchedAmounts(
	ctx context.Context,
	logger log.Logger,
	expectedDenom string,
	channelsToCorrect map[string]string,
	dispatcher *dispatchercomp.Dispatcher,
) error {
	amounts := dispatcher.GetAllDispatchedAmounts(ctx)

	var numDenomUpdated, numChannelUpdated int
	for _, entry := range amounts {
		correctChannel, isWrongChannelID := channelsToCorrect[entry.SourceId.GetCounterpartyId()]

		// The only available route so far is IBC to CCTP. Since CCTP supports only USDC, we have
		// to update all the denoms to USDC.
		isWrongDenom := entry.Denom != expectedDenom

		// If the channel is not in the wrong channel list, or the denom is USDC, then the entry is
		// correct. We basically skip all the entries with 0 incoming dispatched amount but a non
		// zero outgoing amount.
		if !isWrongChannelID && !isWrongDenom {
			// Source protocol is always IBC and destination is always CCTP, no need to log them.
			logger.Debug("skipping dispatched amounts entry",
				"src_counterparty_id", entry.SourceId.GetCounterpartyId(),
				"dst_countertparty_id", entry.DestinationId.GetCounterpartyId(),
				"denom", entry.Denom,
				"amount_incoming", entry.AmountDispatched.Incoming.String(),
				"amount_outgoing", entry.AmountDispatched.Outgoing.String(),
			)
			continue
		}

		// One of the situations to fix, is with a correct channel ID but a wrong denom. Since the
		// correct channel ID is not in the map, we have to use the values of the entry, otherwise
		// it will use the empty string from the miss in the map.
		if !isWrongChannelID {
			correctChannel = entry.SourceId.GetCounterpartyId()
			numDenomUpdated += 1
		} else {
			numChannelUpdated += 1
		}

		logger.Debug("handling dispatched amounts entry",
			"src_counterparty_id", entry.SourceId.GetCounterpartyId(),
			"dst_countertparty_id", entry.DestinationId.GetCounterpartyId(),
			"denom", entry.Denom,
			"amount_incoming", entry.AmountDispatched.Incoming.String(),
			"amount_outgoing", entry.AmountDispatched.Outgoing.String(),
		)

		// We remove from the store the wrong entry.
		err := dispatcher.RemoveDispatchedAmount(ctx, entry.SourceId, entry.DestinationId, entry.Denom)
		if err != nil {
			return fmt.Errorf("failed to remove dispatched amount from state: %w", err)
		}

		correctSourceID, err := orbitercore.NewCrossChainID(entry.SourceId.GetProtocolId(), correctChannel)
		if err != nil {
			return fmt.Errorf("failed to create cross chain ID: %w", err)
		}

		// Get the entry with the correct outgoing value from the state. Returns zero amounts if
		// not present.
		oldValue := dispatcher.GetDispatchedAmount(ctx, &correctSourceID, entry.DestinationId, expectedDenom)

		dispatchedAmount := oldValue.AmountDispatched
		// Add the entry incoming amount as the incoming amount of the correct entry.
		if entry.AmountDispatched.Incoming.IsPositive() {
			dispatchedAmount.Incoming = dispatchedAmount.Incoming.Add(entry.AmountDispatched.Incoming)
		}
		if entry.AmountDispatched.Outgoing.IsPositive() {
			dispatchedAmount.Outgoing = dispatchedAmount.Outgoing.Add(entry.AmountDispatched.Outgoing)
		}

		// Update the entry in the store
		err = dispatcher.SetDispatchedAmount(ctx, &correctSourceID, entry.DestinationId, expectedDenom, dispatchedAmount)
		if err != nil {
			return fmt.Errorf("failed to update the dispatched amount in state: %w", err)
		}
	}

	logger.Info("completed orbiter stats denom update", "updated_entries", numDenomUpdated)
	logger.Info("completed orbiter stats channel update", "updated_entries", numChannelUpdated)
	return nil
}

func updateDispatchedCounts(
	ctx context.Context,
	logger log.Logger,
	channelsToCorrect map[string]string,
	dispatcher *dispatchercomp.Dispatcher,
) error {
	counts := dispatcher.GetAllDispatchedCounts(ctx)

	var numCountsUpdated int
	for _, entry := range counts {
		correctChannel, isWrongChannelID := channelsToCorrect[entry.SourceId.GetCounterpartyId()]

		// If the channel is not wrong, then we don't have to do anything.
		if !isWrongChannelID {
			// Source protocol is always IBC and destination is always CCTP
			logger.Debug("skipping dispatched counts entry",
				"src_counterparty_id", entry.SourceId.GetCounterpartyId(),
				"dst_countertparty_id", entry.DestinationId.GetCounterpartyId(),
				"count", entry.Count,
			)
			continue
		}

		logger.Debug("handling dispatched counts entry",
			"src_counterparty_id", entry.SourceId.GetCounterpartyId(),
			"dst_countertparty_id", entry.DestinationId.GetCounterpartyId(),
			"count", entry.Count,
		)

		// We remove from the store the wrong entry.
		err := dispatcher.RemoveDispatchedCounts(ctx, entry.SourceId, entry.DestinationId)
		if err != nil {
			return fmt.Errorf("failed to remove dispatched counts from state: %w", err)
		}

		correctSourceID, err := orbitercore.NewCrossChainID(entry.SourceId.GetProtocolId(), correctChannel)
		if err != nil {
			return fmt.Errorf("failed to create cross chain ID: %w", err)
		}

		// Get the entry with the correct counts value from the state. Returns zero if not present.
		oldValue := dispatcher.GetDispatchedCounts(ctx, &correctSourceID, entry.DestinationId)

		counts := oldValue.Count + entry.Count

		// Update the entry in the store
		err = dispatcher.SetDispatchedCounts(ctx, &correctSourceID, entry.DestinationId, counts)
		if err != nil {
			return fmt.Errorf("failed to update the dispatched counts in state: %w", err)
		}

		numCountsUpdated += 1
	}
	logger.Info("completed orbiter stats counts update", "updated_entries", numCountsUpdated)

	return nil
}
