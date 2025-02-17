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

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"

	dollarkeeper "dollar.noble.xyz/keeper"
	dollartypes "dollar.noble.xyz/types"
	dollarportaltypes "dollar.noble.xyz/types/portal"

	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	wormholetypes "github.com/noble-assets/wormhole/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	addressCodec address.Codec,
	bankKeeper bankkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	dollarKey *storetypes.KVStoreKey,
	wormholeKeeper *wormholekeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		chainID := sdkCtx.ChainID()
		if chainID != TestnetChainID {
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, chainID)
		}

		// Since M^0 redeployed their entire system on Ethereum Sepolia, we
		// have to reconfigure the entire Noble Dollar state. By deleting the
		// consensus version of the module before RunMigrations, this allows
		// InitGenesis to be rerun. However before migrations, we must first
		// burn all $USDN supply and then clear the entire module state.

		delete(vm, dollartypes.ModuleName)

		err := BurnNobleDollarSupply(ctx, addressCodec, bankKeeper, dollarKeeper)
		if err != nil {
			return vm, err
		}

		err = ClearDollarModuleState(sdkCtx, dollarKey)
		if err != nil {
			return vm, err
		}

		vm, err = mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		err = ConfigureDollarPortalState(ctx, dollarKeeper)
		if err != nil {
			return vm, err
		}

		// Because we removed the GuardianSetExpiry element from the Wormhole
		// configuration, we have to reinitialize the state to avoid running
		// into unmarshalling errors.
		err = ConfigureWormholeState(ctx, wormholeKeeper)
		if err != nil {
			return vm, err
		}

		return vm, nil
	}
}

// BurnNobleDollarSupply burns the entire $USDN supply from the only Noble Dollar user on testnet.
func BurnNobleDollarSupply(ctx context.Context, addressCodec address.Codec, bankKeeper bankkeeper.Keeper, dollarKeeper *dollarkeeper.Keeper) error {
	account, err := addressCodec.StringToBytes("noble1exg6r3tz2tup8ewnmttkkhd7qfx39rguzqngyc")
	if err != nil {
		return errors.Wrap(err, "unable to decode user address")
	}

	denom := dollarKeeper.GetDenom()
	coins := sdk.NewCoins(bankKeeper.GetBalance(ctx, account, denom))

	err = bankKeeper.SendCoinsFromAccountToModule(ctx, account, dollartypes.ModuleName, coins)
	if err != nil {
		return errors.Wrap(err, "failed to transfer usdn to dollar module")
	}
	err = bankKeeper.BurnCoins(ctx, dollartypes.ModuleName, coins)
	if err != nil {
		return errors.Wrap(err, "unable to burn usdn from dollar module")
	}

	supply := bankKeeper.GetSupply(ctx, dollarKeeper.GetDenom())
	if !supply.IsZero() {
		return fmt.Errorf("expected no usdn supply, got %s", supply.Amount)
	}

	return nil
}

// ClearDollarModuleState clears the entire key-value store of the Noble Dollar module.
func ClearDollarModuleState(ctx sdk.Context, dollarKey *storetypes.KVStoreKey) error {
	dollarStore := ctx.KVStore(dollarKey)
	iterator := dollarStore.Iterator(nil, nil)

	for ; iterator.Valid(); iterator.Next() {
		dollarStore.Delete(iterator.Key())
	}

	return iterator.Close()
}

// ConfigureDollarPortalState sets both the Noble Dollar Portal owner and an initial peer.
func ConfigureDollarPortalState(ctx context.Context, dollarKeeper *dollarkeeper.Keeper) (err error) {
	err = dollarKeeper.PortalOwner.Set(ctx, "noble1mx48c5tv6ss9k7793n3a7sv48nfjllhxkd6tq3")
	if err != nil {
		return errors.Wrap(err, "unable to set dollar portal owner in state")
	}

	err = dollarKeeper.PortalPeers.Set(ctx, 10002, dollarportaltypes.Peer{
		// https://sepolia.etherscan.io/address/0xb1725758f7255B025cdbF2814Bc428B403623562
		Transceiver: common.FromHex("0x000000000000000000000000b1725758f7255b025cdbf2814bc428b403623562"),
		// https://sepolia.etherscan.io/address/0xf1669804140fA31cdAA805A1B3Be91e6282D5e41
		Manager: common.FromHex("0x000000000000000000000000f1669804140fa31cdaa805a1b3be91e6282d5e41"),
	})
	if err != nil {
		return errors.Wrap(err, "unable to set dollar portal peer in state")
	}

	return nil
}

// ConfigureWormholeState sets the Wormhole configuration.
func ConfigureWormholeState(ctx context.Context, wormholeKeeper *wormholekeeper.Keeper) (err error) {
	err = wormholeKeeper.Config.Set(ctx, wormholetypes.Config{
		ChainId:          4009,
		GuardianSetIndex: 0,
		GovChain:         1,
		GovAddress:       common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000004"),
	})
	if err != nil {
		return errors.Wrap(err, "unable to set wormhole config in state")
	}

	return nil
}
