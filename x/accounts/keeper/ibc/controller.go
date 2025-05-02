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

package ibc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gogoproto "github.com/cosmos/gogoproto/types"

	"noble.xyz/x/accounts/types"
	"noble.xyz/x/accounts/types/ibc"
)

var _ types.Controller = &Controller{}

type Controller struct {
	cdc codec.Codec
}

func NewController(cdc codec.Codec) Controller {
	return Controller{
		cdc: cdc,
	}
}

func (c Controller) GetAddress(input *gogoproto.Any) ([]byte, error) {
	attributes, err := c.decodeAttributes(input)
	if err != nil {
		return nil, err
	}

	bz := []byte(attributes.Channel + attributes.Recipient + attributes.Fallback)
	return address.Derive([]byte(ibc.ModuleName), bz)[12:], nil
}

func (c Controller) GetSendRestrictionFn() banktypes.SendRestrictionFn {
	return func(ctx context.Context, sender, recipient sdk.AccAddress, coins sdk.Coins) (newRecipient sdk.AccAddress, err error) {
		// TODO: Transfer logic from x/forwarding here!
		return recipient, nil
	}
}

func (c Controller) decodeAttributes(input *gogoproto.Any) (ibc.Attributes, error) {
	var attributes ibc.Attributes
	err := gogoproto.UnmarshalAny(input, &attributes)
	return attributes, err
}
