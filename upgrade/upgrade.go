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
	oriterkeeper "github.com/noble-assets/orbiter/v2/keeper"
	dispatchercomp "github.com/noble-assets/orbiter/v2/keeper/component/dispatcher"
	orbitercore "github.com/noble-assets/orbiter/v2/types/core"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	accountKeeper *authkeeper.AccountKeeper,
	orbiterKeeper *oriterkeeper.Keeper,
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
			return vm, fmt.Errorf("failed to updated Orbiter stats: %w", err)
		}
		writeCache()

		cachedCtx, writeCache = sdkCtx.CacheContext()
		err = updateOrbiterModuleAccounts(cachedCtx, logger, *accountKeeper)
		if err != nil {
			return vm, fmt.Errorf("failed to updated Orbiter module accounts: %w", err)
		}
		writeCache()

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

func updateOrbiterModuleAccounts(ctx sdk.Context, logger log.Logger, accountKeeper authkeeper.AccountKeeper) error {
	for _, name := range []string{orbitercore.ModuleName, orbitercore.DustCollectorName} {

		addr, perms := accountKeeper.GetModuleAddressAndPermissions(name)
		if addr == nil {
			return fmt.Errorf("failed to get module address and permissions for %s", name)
		}

		acc := accountKeeper.GetAccount(ctx, addr)
		baseAcc, ok := (acc).(*authtypes.BaseAccount)
		if !ok {
			_, ok := (acc).(*authtypes.ModuleAccount)
			if ok {
				continue
			}
			return fmt.Errorf("error creating the base account for %s: %T", name, acc)
		}

		macc := authtypes.NewModuleAccount(baseAcc, name, perms...)
		accountKeeper.SetModuleAccount(ctx, macc)
	}

	return nil
}

func updateOrbiterStats(ctx sdk.Context, logger log.Logger, orbiterKeeper *oriterkeeper.Keeper) error {
	expectedDenom := "uusdc"
	var channelsToCorrect map[string]string
	switch ctx.ChainID() {
	case MainnetChainID:
		return nil
	case TestnetChainID:
		// Counterparty channel to Nolbe -> Noble to counterparty channel.
		channelsToCorrect = map[string]string{
			"channel-4280": "channel-22",  // osmosis
			"channel-27":   "channel-639", // namada
			"channel-3":    "channel-333", // ???
			"channel-496":  "channel-43",  // neutron
		}
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
	for _, entry := range amounts {
		correctChannel, isWrongChannelID := channelsToCorrect[entry.SourceId.GetCounterpartyId()]

		// The only available route so far is IBC to CCTP. Since CCTP supports only USDC, we can
		// skip it since it is the correct denom.
		isWrongDenom := entry.Denom == expectedDenom

		if !isWrongChannelID && !isWrongDenom {
			// Source protocol is always IBC and destination is always CCTP
			logger.Debug("skipping dispatched amounts entry",
				"src_counterparty_id", entry.SourceId.GetCounterpartyId(),
				"dst_countertparty_id", entry.DestinationId.GetCounterpartyId(),
				"denom", entry.Denom,
				"amount_incoming", entry.AmountDispatched.Incoming.String(),
				"amount_outgoing", entry.AmountDispatched.Outgoing.String(),
			)
			continue
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

	return nil
}

func updateDispatchedCounts(
	ctx context.Context,
	logger log.Logger,
	channelsToCorrect map[string]string,
	dispatcher *dispatchercomp.Dispatcher,
) error {
	counts := dispatcher.GetAllDispatchedCounts(ctx)
	for _, entry := range counts {
		correctChannel, isWrongChannelID := channelsToCorrect[entry.SourceId.GetCounterpartyId()]

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
	}

	return nil
}
