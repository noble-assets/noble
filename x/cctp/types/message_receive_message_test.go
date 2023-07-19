package types

import (
	"github.com/strangelove-ventures/noble/testutil/sample"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgReceiveMessage_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgReceiveMessage
		err  error
	}{
		{
			name: "happy path",
			msg: MsgReceiveMessage{
				From:        sample.AccAddress(),
				Message:     []byte{1, 2, 3},
				Attestation: []byte{1, 2, 3},
			},
		},
		{
			name: "invalid from",
			msg: MsgReceiveMessage{
				From:        "invalid_address",
				Message:     []byte{1, 2, 3},
				Attestation: []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty message",
			msg: MsgReceiveMessage{
				From:        sample.AccAddress(),
				Message:     []byte{},
				Attestation: []byte{1, 2, 3},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil message",
			msg: MsgReceiveMessage{
				From:        sample.AccAddress(),
				Message:     nil,
				Attestation: []byte{1, 2, 3},
			},
			err: ErrMalformedField,
		},
		{
			name: "empty attestation",
			msg: MsgReceiveMessage{
				From:        sample.AccAddress(),
				Message:     []byte{1, 2, 3},
				Attestation: []byte{},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil attestation",
			msg: MsgReceiveMessage{
				From:        sample.AccAddress(),
				Message:     []byte{},
				Attestation: nil,
			},
			err: ErrMalformedField,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
