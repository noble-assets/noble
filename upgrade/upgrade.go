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

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	clientKeeper clientkeeper.Keeper,
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
			// Substitute the IBC light client for the haqq_11235-1 chain.
			err = clientKeeper.RecoverClient(sdkCtx, "07-tendermint-58", "07-tendermint-194")
			if err != nil {
				logger.Error("failed to recover haqq_11235-1 client", "error", err)
			}
			// Substitute the IBC light client for the migaloo-1 chain.
			err = clientKeeper.RecoverClient(sdkCtx, "07-tendermint-19", "07-tendermint-201")
			if err != nil {
				logger.Error("failed to recover migaloo-1 client", "error", err)
			}
			// Substitute the IBC light client for the omniflixhub-1 chain.
			err = clientKeeper.RecoverClient(sdkCtx, "07-tendermint-68", "07-tendermint-198")
			if err != nil {
				logger.Error("failed to recover omniflixhub-1 client", "error", err)
			}
		}

		return vm, nil
	}
}
