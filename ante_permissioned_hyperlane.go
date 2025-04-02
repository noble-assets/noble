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

	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.AnteDecorator = &PermissionedHyperlaneDecorator{}

// PermissionedHyperlaneDecorator is a custom ante handler that permissions all
// Hyperlane messages on Noble except for a remote transfer on a Warp Route.
type PermissionedHyperlaneDecorator struct {
	cdc codec.Codec
}

func NewPermissionedHyperlaneDecorator(cdc codec.Codec) PermissionedHyperlaneDecorator {
	return PermissionedHyperlaneDecorator{cdc: cdc}
}

func (d PermissionedHyperlaneDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		if _, ok := msg.(*warptypes.MsgRemoteTransfer); ok {
			continue
		}

		typeUrl := sdk.MsgTypeURL(msg)
		if strings.HasPrefix(typeUrl, "/hyperlane") {
			return ctx, fmt.Errorf("%s is currently a permissioned action", typeUrl)
		}
	}

	return next(ctx, tx, simulate)
}
