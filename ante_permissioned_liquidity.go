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

package noble

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	swaptypes "swap.noble.xyz/types"
	stableswaptypes "swap.noble.xyz/types/stableswap"

	"github.com/noble-assets/noble/v11/upgrade"
)

// PermissionedAccount is the account allowed to perform liquidity actions on Noble.
const PermissionedAccount = "noble18vx4czzv4rgrfhm0pzhwu5janjdh4ssdkpu8vr"

var _ sdk.AnteDecorator = &PermissionedLiquidityDecorator{}

// PermissionedLiquidityDecorator is a custom ante handler that permissions all liquidity actions on Noble.
type PermissionedLiquidityDecorator struct{}

func NewPermissionedLiquidityDecorator() PermissionedLiquidityDecorator {
	return PermissionedLiquidityDecorator{}
}

func (d PermissionedLiquidityDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if ctx.ChainID() == upgrade.MainnetChainID {
		for _, msg := range tx.GetMsgs() {
			err := d.CheckMessage(msg)
			if err != nil {
				return ctx, err
			}
		}
	}

	return next(ctx, tx, simulate)
}

func (d PermissionedLiquidityDecorator) CheckMessage(msg sdk.Msg) error {
	switch m := msg.(type) {
	case *stableswaptypes.MsgAddLiquidity:
		if m.Signer != PermissionedAccount {
			return fmt.Errorf("%s is currently a permissioned action", sdk.MsgTypeURL(msg))
		}
	case *stableswaptypes.MsgRemoveLiquidity:
		if m.Signer != PermissionedAccount {
			return fmt.Errorf("%s is currently a permissioned action", sdk.MsgTypeURL(msg))
		}
	case *swaptypes.MsgWithdrawRewards:
		if m.Signer != PermissionedAccount {
			return fmt.Errorf("%s is currently a permissioned action", sdk.MsgTypeURL(msg))
		}
	case *authz.MsgExec:
		execMsgs, err := m.GetMessages()
		if err != nil {
			return err
		}

		for _, execMsg := range execMsgs {
			err = d.CheckMessage(execMsg)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
