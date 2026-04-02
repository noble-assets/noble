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
	"sort"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dollarkeeper "dollar.noble.xyz/v2/keeper"
	vaultstypes "dollar.noble.xyz/v2/types/vaults"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	authoritytypes "github.com/noble-assets/authority/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	addressCdc address.Codec,
	clientKeeper clientkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// In IBC-Go v8.7.0, the MsgRecoverClient message does not support the
		// LegacyAminoJSON signing mode, preventing recovery via the Noble
		// Maintenance Multisig. As a result, expired clients on mainnet must
		// be manually recovered as part of a software upgrade.
		if sdkCtx.ChainID() == MainnetChainID {
			// Substitute the IBC light client for the evmos_9001-2 chain.
			err = clientKeeper.RecoverClient(sdkCtx, "07-tendermint-12", "07-tendermint-208")
			if err != nil {
				logger.Error("failed to recover evmos_9001-2 client", "error", err)
			}
			// Substitute the IBC light client for the haqq_11235-1 chain.
			err = clientKeeper.RecoverClient(sdkCtx, "07-tendermint-58", "07-tendermint-194")
			if err != nil {
				logger.Error("failed to recover haqq_11235-1 client", "error", err)
			}
		}

		vaultServer := dollarkeeper.NewVaultsMsgServer(dollarKeeper)

		// Unlock all user positions from the Staked Vault.
		if err = endVaultsSeasonTwo(ctx, logger, addressCdc, dollarKeeper, vaultServer); err != nil {
			return vm, err
		}

		// Pause all Vaults permanently.
		if _, err = vaultServer.SetPausedState(ctx, &vaultstypes.MsgSetPausedState{
			Signer: authoritytypes.ModuleAddress.String(),
			Paused: vaultstypes.ALL,
		}); err != nil {
			return vm, err
		}

		return vm, nil
	}
}

// endVaultsSeasonTwo handles the logic to end Vaults Season Two, unlocking all
// Staked vault user positions.
func endVaultsSeasonTwo(
	ctx context.Context,
	logger log.Logger,
	addressCdc address.Codec,
	k *dollarkeeper.Keeper,
	vaultServer vaultstypes.MsgServer,
) error {
	// Get all the vaults positions.
	positions, err := k.GetVaultsPositions(ctx)
	if err != nil {
		return err
	}

	// Create a mapping by address and the total positions amount.
	stakedUsers := map[string]math.Int{}
	var stakedUsersAddress []string

	// Iterate through all the positions.
	logger.Info("collecting vault positions")
	for _, position := range positions {
		switch position.Vault {
		case vaultstypes.STAKED:
			addr, err := addressCdc.BytesToString(position.Address)
			if err != nil {
				logger.Warn("invalid position address: " + err.Error())
				continue
			}

			if _, exists := stakedUsers[addr]; !exists {
				stakedUsers[addr] = position.Amount
				stakedUsersAddress = append(stakedUsersAddress, addr)
			} else {
				stakedUsers[addr] = stakedUsers[addr].Add(position.Amount)
			}
		}
	}

	sort.Strings(stakedUsersAddress)

	// Unlock all the Staked vault positions.
	logger.Info(fmt.Sprintf("unlocking %d staked vault positions", len(stakedUsers)))
	stakedUsersProcessed := 0
	for _, stakedUserAddr := range stakedUsersAddress {
		stakedUserTotalAmount := stakedUsers[stakedUserAddr]
		if _, unlockErr := vaultServer.Unlock(ctx, &vaultstypes.MsgUnlock{
			Signer: stakedUserAddr,
			Vault:  vaultstypes.STAKED,
			Amount: stakedUserTotalAmount,
		}); unlockErr != nil {
			logger.Error(fmt.Sprintf("failed to unlock staked vault position for %s: %v", stakedUserAddr, err))
			continue
		}
		stakedUsersProcessed += 1
	}
	logger.Info(fmt.Sprintf("unlocked %d/%d staked vault positions successfully", stakedUsersProcessed, len(stakedUsers)))

	return nil
}
