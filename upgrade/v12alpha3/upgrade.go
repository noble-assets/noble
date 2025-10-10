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

package v12alpha3

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	hyperlanekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/ethereum/go-ethereum/common"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	authoritytypes "github.com/noble-assets/authority/types"
	novakeeper "github.com/noble-assets/nova/keeper"
	novatypes "github.com/noble-assets/nova/types"
	novaismtypes "github.com/noble-assets/nova/types/ism"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	hyperlaneKeeper *hyperlanekeeper.Keeper,
	novaKeeper *novakeeper.Keeper,
	warpKeeper warpkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		chainID := sdk.UnwrapSDKContext(ctx).ChainID()
		if chainID != DevnetChainID {
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, chainID)
		}

		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		if err := clearHyperlaneState(ctx, hyperlaneKeeper); err != nil {
			return vm, fmt.Errorf("failed to clear hyperlane state: %w", err)
		}

		if err := clearWarpState(ctx, warpKeeper); err != nil {
			return vm, fmt.Errorf("failed to clear warp state: %w", err)
		}

		if err := clearNovaState(ctx, novaKeeper); err != nil {
			return vm, fmt.Errorf("failed to clear nova state: %w", err)
		}

		authority, err := authorityKeeper.Owner.Get(ctx)
		if err != nil {
			return vm, fmt.Errorf("unable to get underlying authority address from state: %w", err)
		}
		authorityBz, err := addressCodec.StringToBytes(authority)
		if err != nil {
			return vm, fmt.Errorf("unable to decode underlying authority address: %w", err)
		}

		// Send x/warp module balance back to authority.
		warpAddress := authtypes.NewModuleAddress(warptypes.ModuleName)
		warpBalance := bankKeeper.GetAllBalances(ctx, warpAddress)

		err = bankKeeper.SendCoins(ctx, warpAddress, authorityBz, warpBalance)
		if err != nil {
			return vm, fmt.Errorf("failed to transfer coins from warp module to authority: %w", err)
		}

		// Send $NOBLE balance to x/upgrade module, and then burn.
		users := []string{"noble158asjqdd8ashn9fedw475dwtg05e46kaccc9eg", "noble1hluft9p85fxl8rejst36wsmjq0w0ykzvsduttl"}

		total := sdk.NewCoins()
		for _, user := range users {
			bz, err := addressCodec.StringToBytes(user)
			if err != nil {
				return vm, fmt.Errorf("unable to decode user address %s: %w", user, err)
			}

			balance := bankKeeper.GetBalance(ctx, bz, "anoble")

			err = bankKeeper.SendCoinsFromAccountToModule(ctx, bz, upgradetypes.ModuleName, sdk.NewCoins(balance))
			if err != nil {
				return vm, fmt.Errorf("failed to transfer coins from user %s to upgrade module: %w", user, err)
			}

			total = total.Add(balance)
		}

		err = bankKeeper.BurnCoins(ctx, upgradetypes.ModuleName, total)
		if err != nil {
			return vm, fmt.Errorf("failed to burn coins from upgrade module: %w", err)
		}

		return vm, nil
	}
}

// clearHyperlaneState helps clear the x/hyperlane module state.
func clearHyperlaneState(ctx context.Context, keeper *hyperlanekeeper.Keeper) error {
	hookID, _ := hyperlaneutil.DecodeHexAddress("0x726f757465725f706f73745f6469737061746368000000030000000000000000")
	mailboxID, _ := hyperlaneutil.DecodeHexAddress("0x68797065726c616e650000000000000000000000000000000000000000000000")

	err := keeper.PostDispatchKeeper.SetMerkleTreeHook(ctx, postdispatchtypes.MerkleTreeHook{
		Id:        hookID,
		MailboxId: mailboxID,
		Owner:     authoritytypes.ModuleAddress.String(),
		Tree:      postdispatchtypes.ProtoFromTree(hyperlaneutil.NewTree(hyperlaneutil.ZeroHashes, 0)),
	})
	if err != nil {
		return errors.Wrap(err, "failed to set merkle tree hook")
	}

	mailbox, err := keeper.Mailboxes.Get(ctx, mailboxID.GetInternalId())
	if err != nil {
		return errors.Wrap(err, "unable to get mailbox")
	}
	mailbox.MessageSent = 0
	mailbox.MessageReceived = 0
	if err := keeper.Mailboxes.Set(ctx, mailboxID.GetInternalId(), mailbox); err != nil {
		return errors.Wrap(err, "unable to set mailbox")
	}

	if err := keeper.Messages.Clear(ctx, nil); err != nil {
		return errors.Wrap(err, "unable to clear messages")
	}

	return nil
}

// clearWarpState helps clear the x/warp module state.
func clearWarpState(ctx context.Context, keeper warpkeeper.Keeper) error {
	tokenID, _ := hyperlaneutil.DecodeHexAddress("0x726f757465725f61707000000000000000000000000000010000000000000001")

	usdc, err := keeper.HypTokens.Get(ctx, tokenID.GetInternalId())
	if err != nil {
		return errors.Wrap(err, "unable to get usdc token")
	}
	usdc.CollateralBalance = math.ZeroInt()
	if err := keeper.HypTokens.Set(ctx, tokenID.GetInternalId(), usdc); err != nil {
		return errors.Wrap(err, "unable to set usdc token")
	}

	if err := keeper.EnrolledRouters.Clear(ctx, nil); err != nil {
		return errors.Wrap(err, "unable to clear enrolled routers")
	}

	return nil
}

// clearNovaState helps clear the x/nova module state.
func clearNovaState(ctx context.Context, keeper *novakeeper.Keeper) error {
	genesis := novatypes.GenesisState{
		Ism: novaismtypes.GenesisState{
			Paused: false,
		},
		Config: novatypes.Config{
			EpochLength:        50,
			HookAddress:        common.Address{}.String(),
			EnrolledValidators: []string{"noblevaloper1xynp8wn5r33c6dvhhg3zufn4xghgwz9hsyh3ey"},
		},
		PendingEpoch:    nil,
		FinalizedEpochs: make(map[uint64]novatypes.Epoch),
		StateRoots:      make(map[uint64]string),
		MailboxRoots:    make(map[uint64]string),
	}

	keeper.InitGenesis(ctx, genesis)

	return nil
}
