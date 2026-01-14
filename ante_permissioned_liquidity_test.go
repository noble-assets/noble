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
	swaptypes "swap.noble.xyz/types"
	stableswaptypes "swap.noble.xyz/types/stableswap"

	"github.com/noble-assets/noble/v11/upgrade"
)

func TestPermissionedLiquidityDecorator_CheckMessage(t *testing.T) {
	decorator := NewPermissionedLiquidityDecorator()

	tests := []struct {
		name      string
		msg       sdk.Msg
		expectErr bool
		errMsg    string
	}{
		{
			name: "MsgAddLiquidity with permitted signer",
			msg: &stableswaptypes.MsgAddLiquidity{
				Signer: PermissionedAccount,
			},
			expectErr: false,
		},
		{
			name: "MsgAddLiquidity with unpermitted signer",
			msg: &stableswaptypes.MsgAddLiquidity{
				Signer: "noble1unpermittedsigner123456789",
			},
			expectErr: true,
			errMsg:    "is currently a permissioned action",
		},
		{
			name: "MsgRemoveLiquidity with permitted signer",
			msg: &stableswaptypes.MsgRemoveLiquidity{
				Signer: PermissionedAccount,
			},
			expectErr: false,
		},
		{
			name: "MsgRemoveLiquidity with unpermitted signer",
			msg: &stableswaptypes.MsgRemoveLiquidity{
				Signer: "noble1unpermittedsigner123456789",
			},
			expectErr: true,
			errMsg:    "is currently a permissioned action",
		},
		{
			name: "MsgWithdrawRewards with permitted signer",
			msg: &swaptypes.MsgWithdrawRewards{
				Signer: PermissionedAccount,
			},
			expectErr: false,
		},
		{
			name: "MsgWithdrawRewards with unpermitted signer",
			msg: &swaptypes.MsgWithdrawRewards{
				Signer: "noble1unpermittedsigner123456789",
			},
			expectErr: true,
			errMsg:    "is currently a permissioned action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := decorator.CheckMessage(tt.msg)

			if tt.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPermissionedLiquidityDecorator_CheckMessage_MsgExec(t *testing.T) {
	decorator := NewPermissionedLiquidityDecorator()

	t.Run("MsgExec with nested permitted message", func(t *testing.T) {
		innerMsg := &stableswaptypes.MsgAddLiquidity{
			Signer: PermissionedAccount,
		}
		execMsg := createMsgExec(t, innerMsg)

		err := decorator.CheckMessage(execMsg)
		require.NoError(t, err)
	})

	t.Run("MsgExec with nested unpermitted message", func(t *testing.T) {
		innerMsg := &stableswaptypes.MsgAddLiquidity{
			Signer: "noble1unpermittedsigner123456789",
		}
		execMsg := createMsgExec(t, innerMsg)

		err := decorator.CheckMessage(execMsg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is currently a permissioned action")
	})
}

func TestPermissionedLiquidityDecorator_AnteHandle_MainnetOnly(t *testing.T) {
	decorator := NewPermissionedLiquidityDecorator()

	nextCalled := false
	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	}

	t.Run("mainnet - unpermitted signer should fail", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: upgrade.MainnetChainID}, false, nil)
		tx := &mockTx{
			msgs: []sdk.Msg{
				&stableswaptypes.MsgAddLiquidity{
					Signer: "noble1unpermittedsigner123456789",
				},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.Error(t, err)
		require.False(t, nextCalled)
	})

	t.Run("mainnet - permitted signer should pass", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: upgrade.MainnetChainID}, false, nil)
		tx := &mockTx{
			msgs: []sdk.Msg{
				&stableswaptypes.MsgAddLiquidity{
					Signer: PermissionedAccount,
				},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.True(t, nextCalled)
	})

	t.Run("testnet - any signer should pass", func(t *testing.T) {
		nextCalled = false
		ctx := sdk.NewContext(nil, cmtproto.Header{ChainID: "grand-1"}, false, nil)
		tx := &mockTx{
			msgs: []sdk.Msg{
				&stableswaptypes.MsgAddLiquidity{
					Signer: "noble1unpermittedsigner123456789",
				},
			},
		}

		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.True(t, nextCalled)
	})
}

func TestPermissionedLiquidityDecorator_CheckMessage_UnrelatedMessage(t *testing.T) {
	decorator := NewPermissionedLiquidityDecorator()

	// Unrelated message types should pass through without error
	t.Run("unrelated message type should pass", func(t *testing.T) {
		msg := &mockUnrelatedMsg{}
		err := decorator.CheckMessage(msg)
		require.NoError(t, err)
	})
}

// Helper function to create MsgExec with inner message
func createMsgExec(t *testing.T, innerMsg sdk.Msg) *authz.MsgExec {
	t.Helper()

	anyMsg, err := types.NewAnyWithValue(innerMsg)
	require.NoError(t, err)

	return &authz.MsgExec{
		Grantee: "noble1grantee",
		Msgs:    []*types.Any{anyMsg},
	}
}

// Mock transaction for testing
type mockTx struct {
	msgs []sdk.Msg
}

func (m *mockTx) GetMsgs() []sdk.Msg {
	return m.msgs
}

func (m *mockTx) GetMsgsV2() ([]proto.Message, error) {
	protoMsgs := make([]proto.Message, len(m.msgs))
	for i, msg := range m.msgs {
		protoMsgs[i] = msg.(proto.Message)
	}
	return protoMsgs, nil
}

// Mock unrelated message for testing passthrough behavior
type mockUnrelatedMsg struct{}

func (m *mockUnrelatedMsg) Reset()         {}
func (m *mockUnrelatedMsg) String() string { return "mockUnrelatedMsg" }
func (m *mockUnrelatedMsg) ProtoMessage()  {}
