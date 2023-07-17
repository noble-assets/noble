package types

import (
	"testing"

	"github.com/strangelove-ventures/noble/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgSendMessage_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSendMessage
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgSendMessage{
				From:              "invalid_address",
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{2, 3, 4},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid from",
			msg: MsgSendMessage{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{2, 3, 4},
			},
		},
		{
			name: "nil recipient",
			msg: MsgSendMessage{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         nil,
				MessageBody:       []byte{2, 3, 4},
			},
			err: ErrMalformedField,
		},
		{
			name: "empty recipient",
			msg: MsgSendMessage{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{},
				MessageBody:       []byte{2, 3, 4},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil message body",
			msg: MsgSendMessage{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{1, 2, 3},
				MessageBody:       nil,
			},
			err: ErrMalformedField,
		},
		{
			name: "empty message body",
			msg: MsgSendMessage{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{1, 2, 3},
				MessageBody:       []byte{},
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
