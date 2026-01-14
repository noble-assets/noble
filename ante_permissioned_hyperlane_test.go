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
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	hyperlanetypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
)

// mockDollarKeeper implements the DollarKeeper interface for testing
type mockDollarKeeper struct {
	denom string
}

func (m *mockDollarKeeper) GetDenom() string {
	return m.denom
}

func TestPermissionedHyperlaneDecorator_CheckMessage_AllowedMessages(t *testing.T) {
	dollarKeeper := &mockDollarKeeper{denom: "uusdn"}
	decorator := NewPermissionedHyperlaneDecorator(dollarKeeper)

	tests := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgAnnounceValidator is allowed",
			msg:  &ismtypes.MsgAnnounceValidator{},
		},
		{
			name: "MsgProcessMessage is allowed",
			msg:  &hyperlanetypes.MsgProcessMessage{},
		},
		{
			name: "MsgSetToken is allowed",
			msg:  &warptypes.MsgSetToken{},
		},
		{
			name: "MsgEnrollRemoteRouter is allowed",
			msg:  &warptypes.MsgEnrollRemoteRouter{},
		},
		{
			name: "MsgUnrollRemoteRouter is allowed",
			msg:  &warptypes.MsgUnrollRemoteRouter{},
		},
		{
			name: "MsgRemoteTransfer is allowed",
			msg:  &warptypes.MsgRemoteTransfer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decorator.CheckMessage(tt.msg)
			require.NoError(t, err)
		})
	}
}

func TestPermissionedHyperlaneDecorator_CheckMessage_CreateCollateralToken(t *testing.T) {
	dollarDenom := "uusdn"
	dollarKeeper := &mockDollarKeeper{denom: dollarDenom}
	decorator := NewPermissionedHyperlaneDecorator(dollarKeeper)

	tests := []struct {
		name        string
		originDenom string
		expectErr   bool
		errMsg      string
	}{
		{
			name:        "dollar denom is restricted",
			originDenom: dollarDenom,
			expectErr:   true,
			errMsg:      "cannot create hyperlane collateral token for denom",
		},
		{
			name:        "other denom is allowed",
			originDenom: "uusdc",
			expectErr:   false,
		},
		{
			name:        "empty denom is allowed",
			originDenom: "",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &warptypes.MsgCreateCollateralToken{
				OriginDenom: tt.originDenom,
			}

			err := decorator.CheckMessage(msg)

			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPermissionedHyperlaneDecorator_CheckMessage_MsgExec(t *testing.T) {
	dollarKeeper := &mockDollarKeeper{denom: "uusdn"}
	decorator := NewPermissionedHyperlaneDecorator(dollarKeeper)

	t.Run("MsgExec with allowed nested message", func(t *testing.T) {
		innerMsg := &ismtypes.MsgAnnounceValidator{}
		execMsg := createHyperlaneMsgExec(t, innerMsg)

		err := decorator.CheckMessage(execMsg)
		require.NoError(t, err)
	})

	t.Run("MsgExec with restricted nested message", func(t *testing.T) {
		innerMsg := &warptypes.MsgCreateCollateralToken{
			OriginDenom: "uusdn",
		}
		execMsg := createHyperlaneMsgExec(t, innerMsg)

		err := decorator.CheckMessage(execMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot create hyperlane collateral token")
	})
}

func TestPermissionedHyperlaneDecorator_AnteHandle(t *testing.T) {
	dollarKeeper := &mockDollarKeeper{denom: "uusdn"}
	decorator := NewPermissionedHyperlaneDecorator(dollarKeeper)

	nextCalled := false
	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	}

	t.Run("allowed message passes through", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: "noble-1"}, false, nil)
		tx := &mockHyperlaneTx{
			msgs: []sdk.Msg{
				&ismtypes.MsgAnnounceValidator{},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.True(t, nextCalled)
	})

	t.Run("restricted message fails", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: "noble-1"}, false, nil)
		tx := &mockHyperlaneTx{
			msgs: []sdk.Msg{
				&warptypes.MsgCreateCollateralToken{
					OriginDenom: "uusdn",
				},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.Error(t, err)
		require.False(t, nextCalled)
	})

	t.Run("multiple messages - all allowed", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: "noble-1"}, false, nil)
		tx := &mockHyperlaneTx{
			msgs: []sdk.Msg{
				&ismtypes.MsgAnnounceValidator{},
				&warptypes.MsgSetToken{},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.True(t, nextCalled)
	})

	t.Run("multiple messages - one restricted", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: "noble-1"}, false, nil)
		tx := &mockHyperlaneTx{
			msgs: []sdk.Msg{
				&ismtypes.MsgAnnounceValidator{},
				&warptypes.MsgCreateCollateralToken{
					OriginDenom: "uusdn",
				},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.Error(t, err)
		require.False(t, nextCalled)
	})
}

func TestPermissionedHyperlaneDecorator_CheckMessage_UnrelatedMessage(t *testing.T) {
	dollarKeeper := &mockDollarKeeper{denom: "uusdn"}
	decorator := NewPermissionedHyperlaneDecorator(dollarKeeper)

	// Unrelated message types should pass through without error
	t.Run("unrelated message type should pass", func(t *testing.T) {
		msg := &mockHyperlaneUnrelatedMsg{}
		err := decorator.CheckMessage(msg)
		require.NoError(t, err)
	})
}

// Helper function to create MsgExec with inner message for Hyperlane tests
func createHyperlaneMsgExec(t *testing.T, innerMsg sdk.Msg) *authz.MsgExec {
	t.Helper()

	anyMsg, err := types.NewAnyWithValue(innerMsg)
	require.NoError(t, err)

	return &authz.MsgExec{
		Grantee: "noble1grantee",
		Msgs:    []*types.Any{anyMsg},
	}
}

// Mock transaction for Hyperlane testing
type mockHyperlaneTx struct {
	msgs []sdk.Msg
}

func (m *mockHyperlaneTx) GetMsgs() []sdk.Msg {
	return m.msgs
}

func (m *mockHyperlaneTx) GetMsgsV2() ([]proto.Message, error) {
	protoMsgs := make([]proto.Message, len(m.msgs))
	for i, msg := range m.msgs {
		protoMsgs[i] = msg.(proto.Message)
	}
	return protoMsgs, nil
}

// Mock unrelated message for testing passthrough behavior
type mockHyperlaneUnrelatedMsg struct{}

func (m *mockHyperlaneUnrelatedMsg) Reset()         {}
func (m *mockHyperlaneUnrelatedMsg) String() string { return "mockHyperlaneUnrelatedMsg" }
func (m *mockHyperlaneUnrelatedMsg) ProtoMessage()  {}
