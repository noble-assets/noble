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

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/ethereum/go-ethereum/common"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	wormholetypes "github.com/noble-assets/wormhole/types"

	dollarkeeper "dollar.noble.xyz/keeper"
	portaltypes "dollar.noble.xyz/types/portal"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	capabilityKeeper *capabilitykeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	wormholeKeeper *wormholekeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)

		FixICS27ChannelCapabilities(sdkCtx, capabilityKeeper)

		if err := ConfigureWormholeModule(sdkCtx, wormholeKeeper); err != nil {
			return vm, err
		}

		if err := ConfigureDollarModule(sdkCtx, dollarKeeper); err != nil {
			return vm, err
		}

		logger.Info("Welcome to a new generation of Noble!" + UpgradeASCII)

		return vm, nil
	}
}

// FixICS27ChannelCapabilities finds all capabilities wrongfully owned by the
// ICA Controller module and replaces them with the ICA Host module. This was
// introduced in the v8 Helium upgrade after we executed the recommended ICS27
// migration logic for chains that utilize the ICA Controller module.
func FixICS27ChannelCapabilities(ctx sdk.Context, capabilityKeeper *capabilitykeeper.Keeper) {
	index := capabilityKeeper.GetLatestIndex(ctx)

	for i := uint64(1); i < index; i++ {
		wrapper, ok := capabilityKeeper.GetOwners(ctx, i)
		if !ok {
			continue
		}

		for _, owner := range wrapper.GetOwners() {
			if owner.Module == icacontrollertypes.SubModuleName {
				wrapper.Remove(owner)
				wrapper.Set(capabilitytypes.Owner{
					Module: icahosttypes.SubModuleName,
					Name:   owner.Name,
				})
			}
		}

		capabilityKeeper.SetOwners(ctx, i, wrapper)
	}

	capabilityKeeper.InitMemStore(ctx)
}

// ConfigureWormholeModule sets both the Wormhole module configuration and an initial guardian set.
func ConfigureWormholeModule(ctx sdk.Context, wormholeKeeper *wormholekeeper.Keeper) (err error) {
	switch ctx.ChainID() {
	case TestnetChainID:
		err = wormholeKeeper.Config.Set(ctx, wormholetypes.Config{
			ChainId:           4009,
			GuardianSetIndex:  0,
			GuardianSetExpiry: 0,
			GovChain:          1,
			GovAddress:        common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000004"),
		})
		if err != nil {
			return errors.Wrap(err, "unable to set wormhole config in state")
		}

		err = wormholeKeeper.GuardianSets.Set(ctx, 0, wormholetypes.GuardianSet{
			// https://github.com/wormhole-foundation/wormhole/blob/3797ed082150e6d66c0dce3fea7f2848364af7d5/ethereum/env/.env.sepolia.testnet#L7
			Addresses:      [][]byte{common.FromHex("0x13947Bd48b18E53fdAeEe77F3473391aC727C638")},
			ExpirationTime: 0,
		})
		if err != nil {
			return errors.Wrap(err, "unable to set wormhole guardian set in state")
		}

		return nil
	case MainnetChainID:
		// TODO: Add the necessary configurations for mainnet here!
		return nil
	default:
		return fmt.Errorf("cannot configure the wormhole module on %s chain", ctx.ChainID())
	}
}

// ConfigureDollarModule sets both the Dollar Portal submodule owner and an initial peer.
func ConfigureDollarModule(ctx sdk.Context, dollarKeeper *dollarkeeper.Keeper) (err error) {
	switch ctx.ChainID() {
	case TestnetChainID:
		err = dollarKeeper.Owner.Set(ctx, "noble1mx48c5tv6ss9k7793n3a7sv48nfjllhxkd6tq3")
		if err != nil {
			return errors.Wrap(err, "unable to set dollar portal owner in state")
		}

		err = dollarKeeper.Peers.Set(ctx, 10002, portaltypes.Peer{
			// https://sepolia.etherscan.io/address/0x29CbF1e07166D31446307aE07999fa6d16223990
			Transceiver: common.FromHex("0x00000000000000000000000029cbf1e07166d31446307ae07999fa6d16223990"),
			// https://sepolia.etherscan.io/address/0x1B7aE194B20C555B9d999c835F74cDCE36A67a74
			Manager: common.FromHex("0x0000000000000000000000001b7ae194b20c555b9d999c835f74cdce36a67a74"),
		})
		if err != nil {
			return errors.Wrap(err, "unable to set dollar portal peer in state")
		}

		return nil
	case MainnetChainID:
		// TODO: Add the necessary configurations for mainnet here!
		return nil
	default:
		return fmt.Errorf("cannot configure the dollar module on %s chain", ctx.ChainID())
	}
}
