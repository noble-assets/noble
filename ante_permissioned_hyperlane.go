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
	"strings"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	hyperlanetypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ sdk.AnteDecorator = &PermissionedHyperlaneDecorator{}

// PermissionedHyperlaneDecorator is a custom ante handler that permissions all Hyperlane actions on Noble.
type PermissionedHyperlaneDecorator struct {
	dollarKeeper DollarKeeper
}

// DollarKeeper defines the interface expected by PermissionedHyperlaneDecorator for the Noble Dollar module.
type DollarKeeper interface {
	GetDenom() string
}

func NewPermissionedHyperlaneDecorator(dollarKeeper DollarKeeper) PermissionedHyperlaneDecorator {
	return PermissionedHyperlaneDecorator{
		dollarKeeper: dollarKeeper,
	}
}

func (d PermissionedHyperlaneDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		err := d.CheckMessage(msg)
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

func (d PermissionedHyperlaneDecorator) CheckMessage(msg sdk.Msg) error {
	switch m := msg.(type) {
	case *ismtypes.MsgAnnounceValidator:
		return nil
	case *hyperlanetypes.MsgProcessMessage:
		return nil
	case *warptypes.MsgCreateCollateralToken:
		if m.OriginDenom == d.dollarKeeper.GetDenom() {
			return fmt.Errorf("cannot create hyperlane collateral token for denom %s", m.OriginDenom)
		}

		return nil
	case *warptypes.MsgRemoteTransfer:
		return nil
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
	default:
		typeUrl := sdk.MsgTypeURL(msg)
		if strings.HasPrefix(typeUrl, "/hyperlane") {
			if strings.HasPrefix(typeUrl, "/hyperlane.core.post_dispatch") {
				return nil
			}

			return fmt.Errorf("%s is currently a permissioned action", typeUrl)
		}
	}

	return nil
}
