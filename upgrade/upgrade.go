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

	"cosmossdk.io/collections"
	"cosmossdk.io/core/header"
	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	wormholetypes "github.com/noble-assets/wormhole/types"
	vaautils "github.com/wormhole-foundation/wormhole/sdk/vaa"

	dollarkeeper "dollar.noble.xyz/keeper"
	portaltypes "dollar.noble.xyz/types/portal"

	swapkeeper "swap.noble.xyz/keeper"
	swaptypes "swap.noble.xyz/types"
	stableswaptypes "swap.noble.xyz/types/stableswap"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	capabilityKeeper *capabilitykeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	wormholeKeeper *wormholekeeper.Keeper,
	swapKeeper *swapkeeper.Keeper,
	accountKeeper keeper.AccountKeeper,
	headerService header.Service,
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

		if err := ConfigureSwapModule(sdkCtx, swapKeeper, accountKeeper, headerService); err != nil {
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

// ConfigureWormholeModule sets both the configuration and an initial guardian set for Wormhole.
func ConfigureWormholeModule(ctx sdk.Context, wormholeKeeper *wormholekeeper.Keeper) (err error) {
	err = wormholeKeeper.Config.Set(ctx, wormholetypes.Config{
		ChainId:          uint16(vaautils.ChainIDNoble),
		GuardianSetIndex: 0,
		GovChain:         uint16(vaautils.GovernanceChain),
		GovAddress:       vaautils.GovernanceEmitter.Bytes(),
	})
	if err != nil {
		return errors.Wrap(err, "unable to set wormhole config in state")
	}

	switch ctx.ChainID() {
	case TestnetChainID:
		err = wormholeKeeper.GuardianSets.Set(ctx, 0, wormholetypes.GuardianSet{
			// https://github.com/wormhole-foundation/wormhole/blob/3797ed082150e6d66c0dce3fea7f2848364af7d5/ethereum/env/.env.sepolia.testnet#L7
			Addresses:      [][]byte{common.FromHex("0x13947Bd48b18E53fdAeEe77F3473391aC727C638")},
			ExpirationTime: 0,
		})
		if err != nil {
			return errors.Wrap(err, "unable to set initial wormhole guardian set in state")
		}

		return nil
	case MainnetChainID:
		err = wormholeKeeper.GuardianSets.Set(ctx, 0, wormholetypes.GuardianSet{
			// https://github.com/wormhole-foundation/wormhole/blob/3797ed082150e6d66c0dce3fea7f2848364af7d5/ethereum/env/.env.ethereum.mainnet#L4
			Addresses:      [][]byte{common.FromHex("0x58CC3AE5C097b213cE3c81979e1B9f9570746AA5")},
			ExpirationTime: 0,
		})
		if err != nil {
			return errors.Wrap(err, "unable to set initial wormhole guardian set in state")
		}

		return nil
	default:
		return fmt.Errorf("cannot configure initial wormhole guardian set on %s chain", ctx.ChainID())
	}
}

// ConfigureDollarModule sets the owner, a peer, and supported bridging paths for the Noble Dollar Portal.
func ConfigureDollarModule(ctx sdk.Context, dollarKeeper *dollarkeeper.Keeper) (err error) {
	// NOTE: The $M token address is the same across all EVM networks.
	//
	// https://etherscan.io/address/0x866A2BF4E572CbcF37D5071A7a58503Bfb36be1b
	// https://sepolia.etherscan.io/address/0x866A2BF4E572CbcF37D5071A7a58503Bfb36be1b
	m := common.FromHex("0x000000000000000000000000866a2bf4e572cbcf37d5071a7a58503bfb36be1b")
	// NOTE: The $wM token address is the same across all EVM networks.
	//
	// https://etherscan.io/address/0x437cc33344a0B27A429f795ff6B469C72698B291
	// https://sepolia.etherscan.io/address/0x437cc33344a0B27A429f795ff6B469C72698B291
	wm := common.FromHex("0x000000000000000000000000437cc33344a0b27a429f795ff6b469c72698b291")

	switch ctx.ChainID() {
	case TestnetChainID:
		chainID := uint16(vaautils.ChainIDSepolia)

		err = dollarKeeper.PortalOwner.Set(ctx, "noble1mx48c5tv6ss9k7793n3a7sv48nfjllhxkd6tq3")
		if err != nil {
			return errors.Wrap(err, "unable to set dollar portal owner in state")
		}

		err = dollarKeeper.PortalPeers.Set(ctx, chainID, portaltypes.Peer{
			// https://sepolia.etherscan.io/address/0x0763196A091575adF99e2306E5e90E0Be5154841
			Transceiver: common.FromHex("0x0000000000000000000000000763196a091575adf99e2306e5e90e0be5154841"),
			// https://sepolia.etherscan.io/address/0xD925C84b55E4e44a53749fF5F2a5A13F63D128fd
			Manager: common.FromHex("0x000000000000000000000000d925c84b55e4e44a53749ff5f2a5a13f63d128fd"),
		})
		if err != nil {
			return errors.Wrap(err, "unable to set dollar portal peer in state")
		}

		// $USDN -> $M
		err = dollarKeeper.PortalBridgingPaths.Set(ctx, collections.Join(chainID, m), true)
		if err != nil {
			return errors.Wrap(err, "unable to set first dollar portal bridging path in state")
		}
		// $USDN -> $wM
		err = dollarKeeper.PortalBridgingPaths.Set(ctx, collections.Join(chainID, wm), true)
		if err != nil {
			return errors.Wrap(err, "unable to set second dollar portal bridging path in state")
		}

		return nil
	case MainnetChainID:
		chainID := uint16(vaautils.ChainIDEthereum)

		// TODO: Add portal owner configuration!

		err = dollarKeeper.PortalPeers.Set(ctx, chainID, portaltypes.Peer{
			// TODO: Confirm Noble's mainnnet transceiver address with M^0

			// https://etherscan.io/address/0x83Ae82Bd4054e815fB7B189C39D9CE670369ea16
			Manager: common.FromHex("0x00000000000000000000000083ae82bd4054e815fb7b189c39d9ce670369ea16"),
		})

		// $USDN -> $M
		err = dollarKeeper.PortalBridgingPaths.Set(ctx, collections.Join(chainID, m), true)
		if err != nil {
			return errors.Wrap(err, "unable to set first dollar portal bridging path in state")
		}
		// $USDN -> $wM
		err = dollarKeeper.PortalBridgingPaths.Set(ctx, collections.Join(chainID, wm), true)
		if err != nil {
			return errors.Wrap(err, "unable to set second dollar portal bridging path in state")
		}

		return nil
	default:
		return fmt.Errorf("cannot configure the dollar portal on %s chain", ctx.ChainID())
	}
}

// ConfigureSwapModule initializes the initial Swap Module state.
func ConfigureSwapModule(ctx sdk.Context, swapKeeper *swapkeeper.Keeper, accountKeeper keeper.AccountKeeper, headerService header.Service) (err error) {
	switch ctx.ChainID() {
	case MainnetChainID:
		poolId, err := swapKeeper.IncreaseNextPoolID(ctx)
		if err != nil {
			return errors.Wrapf(err, "unable to set next pool id")
		}

		// Create the Pool address.
		prefix := fmt.Sprintf("%s/pool/%d", swaptypes.ModuleName, poolId)
		account := authtypes.NewEmptyModuleAccount(prefix)
		account = accountKeeper.NewAccount(ctx, account).(*authtypes.ModuleAccount)
		accountKeeper.SetModuleAccount(ctx, account)

		// Create the Protocol Fees Pool address.
		protocolFeesAccount := authtypes.NewEmptyModuleAccount(fmt.Sprintf("%s/protocol_fees", prefix))
		protocolFees := accountKeeper.NewAccount(ctx, protocolFeesAccount).(*authtypes.ModuleAccount)
		accountKeeper.SetModuleAccount(ctx, protocolFees)

		// Create the Rewards Fees Pool address.
		rewardFeesAccount := authtypes.NewEmptyModuleAccount(fmt.Sprintf("%s/reward_fees", prefix))
		rewardFees := accountKeeper.NewAccount(ctx, rewardFeesAccount).(*authtypes.ModuleAccount)
		accountKeeper.SetModuleAccount(ctx, rewardFees)
		err = swapKeeper.SetPool(ctx, 0, swaptypes.Pool{
			Id:        poolId,
			Address:   account.GetAddress().String(),
			Algorithm: swaptypes.STABLESWAP,
			Pair:      "uusdc",
		})
		if err != nil {
			return errors.Wrap(err, "unable to set initial swap pool in state")
		}

		// Add the new Pool ID to the `Paused` state.
		if err = swapKeeper.SetPaused(ctx, poolId, false); err != nil {
			return errors.Wrapf(err, "unable to create paused pool initial entry")
		}

		err = swapKeeper.Stableswap.SetPool(ctx, 0, stableswaptypes.Pool{
			ProtocolFeePercentage: 50,      //TODO: wait final params
			RewardsFee:            2500000, //TODO: wait final params
			InitialA:              800,     //TODO: wait final params
			FutureA:               800,     //TODO: wait final params
			InitialATime:          headerService.GetHeaderInfo(ctx).Time.Unix(),
			FutureATime:           0, //TODO: wait final params
			RateMultipliers: sdk.NewCoins(
				sdk.NewCoin("uusdn", math.NewInt(1e18)),
				sdk.NewCoin("uusdc", math.NewInt(1e18)),
			),
			TotalShares: math.LegacyZeroDec(),
		})
		if err != nil {
			return errors.Wrap(err, "unable to set initial swap pool in state")
		}

		return nil
	}

	return nil
}
