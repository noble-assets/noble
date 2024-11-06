// Copyright 2024 NASD Inc. All Rights Reserved.
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

package testnet

import (
	"context"
	"sort"
	"strings"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	globalfeekeeper "github.com/noble-assets/globalfee/keeper"
	globalfeetypes "github.com/noble-assets/globalfee/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	registry codectypes.InterfaceRegistry,
	globalFeeKeeper *globalfeekeeper.Keeper,
	paramsKeeper paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Initialize the legacy GlobalFee param subspace to enable migration.
		subspace, _ := paramsKeeper.GetSubspace(globalfeetypes.ModuleName)
		subspace.WithKeyTable(globalfeetypes.ParamKeyTable()) //nolint:staticcheck

		// Migrate GlobalFee legacy params to state of the new standalone module.
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		// Override migrated list of bypass messages, ensuring that IBC relaying
		// remains free, and enable all current asset issuers (Circle, Ondo,
		// Hashnote, and Monerium) to interact with the protocol for free.
		bypassMessages := []string{
			sdk.MsgTypeURL(&clienttypes.MsgUpdateClient{}),
			sdk.MsgTypeURL(&channeltypes.MsgRecvPacket{}),
			sdk.MsgTypeURL(&channeltypes.MsgTimeout{}),
			sdk.MsgTypeURL(&channeltypes.MsgAcknowledgement{}),
		}
		bypassMessages = append(bypassMessages, GetModuleMessages(registry, "circle")...)
		bypassMessages = append(bypassMessages, GetModuleMessages(registry, "aura")...)
		bypassMessages = append(bypassMessages, GetModuleMessages(registry, "halo")...)
		bypassMessages = append(bypassMessages, GetModuleMessages(registry, "florin")...)
		sort.Strings(bypassMessages)

		err = globalFeeKeeper.BypassMessages.Clear(ctx, nil)
		if err != nil {
			return vm, err
		}
		for _, bypassMessage := range bypassMessages {
			err = globalFeeKeeper.BypassMessages.Set(ctx, bypassMessage)
			if err != nil {
				return vm, err
			}
		}

		return vm, nil
	}
}

// GetModuleMessages is a utility that returns all messages registered by a module.
func GetModuleMessages(registry codectypes.InterfaceRegistry, name string) (messages []string) {
	for _, message := range registry.ListImplementations(sdk.MsgInterfaceProtoName) {
		if strings.HasPrefix(message, "/"+name) {
			messages = append(messages, message)
		}
	}

	sort.Strings(messages)
	return
}
