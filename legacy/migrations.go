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

package legacy

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"

	dollarkeeper "dollar.noble.xyz/v2/keeper"
	dollartypes "dollar.noble.xyz/v2/types"
	dollartypesv2 "dollar.noble.xyz/v2/types/v2"
)

// MigrateDollar is a temporary migration for Noble's testnet (grand-1) to be
// performed at Block 30,446,659 that migrates the state of the Noble Dollar
// module from v2.0.0-rc.1 to v2.0.0-rc.2
func MigrateDollar(ctx context.Context, cdc codec.BinaryCodec, service store.KVStoreService, keeper *dollarkeeper.Keeper) error {
	builder := collections.NewSchemaBuilder(service)
	legacyCollection := collections.NewItem(builder, dollartypes.StatsKey, "stats", codec.CollValue[Stats](cdc))
	if _, err := builder.Build(); err != nil {
		return err
	}

	legacyStats, err := legacyCollection.Get(ctx)
	if err != nil {
		return err
	}

	err = keeper.Stats.Set(ctx, dollartypesv2.Stats{
		TotalHolders:      legacyStats.TotalHolders,
		TotalPrincipal:    legacyStats.TotalPrincipal,
		TotalYieldAccrued: legacyStats.TotalYieldAccrued,
	})
	if err != nil {
		return err
	}

	for key, rawAmount := range legacyStats.TotalExternalYield {
		provider, identier := dollartypesv2.ParseYieldRecipientKey(key)
		amount, _ := math.NewIntFromString(rawAmount)

		err = keeper.TotalExternalYield.Set(
			ctx,
			collections.Join(int32(provider), identier),
			amount,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
