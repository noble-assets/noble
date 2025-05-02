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

package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"noble.xyz/x/accounts/keeper/cctp"
	"noble.xyz/x/accounts/keeper/hyperlane"
	"noble.xyz/x/accounts/keeper/ibc"
	"noble.xyz/x/accounts/types"
)

type Keeper struct {
	controllers map[string]types.Controller
}

func NewKeeper(cdc codec.Codec) Keeper {
	controllers := make(map[string]types.Controller)

	cctpController := cctp.NewController(cdc)
	controllers["cctp"] = cctpController
	hyperlaneController := hyperlane.NewController(cdc)
	controllers["hyperlane"] = hyperlaneController
	ibcController := ibc.NewController(cdc)
	controllers["ibc"] = ibcController

	return Keeper{
		controllers: controllers,
	}
}

func (k *Keeper) SendRestrictionFn(ctx context.Context, sender, recipient sdk.AccAddress, coins sdk.Coins) (newRecipient sdk.AccAddress, err error) {
	for _, controller := range k.controllers {
		recipient, err = controller.GetSendRestrictionFn()(ctx, sender, recipient, coins)
		if err != nil {
			return recipient, err
		}
	}

	return recipient, nil
}
