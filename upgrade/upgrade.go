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

	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	ismkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authoritytypes "github.com/noble-assets/authority/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	ismKeeper ismkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		// Because the HyperEVM default ISM has already been created on Noble's
		// testnet, we only need to run the logic for creating the HyperEVM
		// default ISM on Noble's mainnet.
		if sdkCtx.ChainID() == MainnetChainID {
			err = configureHyperlaneISM(ctx, logger, ismKeeper)
			if err != nil {
				return vm, err
			}
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

func configureHyperlaneISM(ctx context.Context, logger log.Logger, keeper ismkeeper.Keeper) error {
	// https://docs.hyperlane.xyz/docs/reference/addresses/validators/mainnet-default-ism-validators#hyperevm-999
	domain := uint32(999)
	validators := []string{
		"0x01be14a9eceeca36c9c1d46c056ca8c87f77c26f", // Abacus Works
		"0xcf0211fafbb91fd9d06d7e306b30032dc3a1934f", // Merkly
		"0x4f977a59fdc2d9e39f6d780a84d5b4add1495a36", // Mitosis
		"0x04d949c615c9976f89595ddcb9008c92f8ba7278", // Luganodes
	}
	sort.Strings(validators)
	threshold := uint32(3)

	id, err := keeper.CreateMerkleRootMultisigIsm(ctx, &ismtypes.MsgCreateMerkleRootMultisigIsm{
		Creator:    authoritytypes.ModuleAddress.String(),
		Validators: validators,
		Threshold:  threshold,
	})
	if err != nil {
		return fmt.Errorf("unable to create default ism for domain %d: %w", domain, err)
	}

	logger.Info("created default hyperlane ism for hyperevm", "domain", domain, "id", id)

	err = keeper.SetRoutingIsmDomain(ctx, &ismtypes.MsgSetRoutingIsmDomain{
		IsmId: DefaultISM,
		Route: ismtypes.Route{
			Ism:    id,
			Domain: domain,
		},
		Owner: authoritytypes.ModuleAddress.String(),
	})
	if err != nil {
		return fmt.Errorf("unable to set default ism in routing ism: %w", err)
	}

	return nil
}
