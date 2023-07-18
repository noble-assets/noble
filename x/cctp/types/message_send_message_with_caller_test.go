package types

import (
	"github.com/strangelove-ventures/noble/testutil/sample"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgSendMessageWithCaller_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSendMessageWithCaller
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgSendMessageWithCaller{
				From:              "invalid_address",
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{2, 3, 4},
				DestinationCaller: []byte{2, 3, 4},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid from",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{2, 3, 4},
				DestinationCaller: []byte{2, 3, 4},
			},
		},
		{
			name: "empty recipient",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{},
				MessageBody:       []byte{2, 3, 4},
				DestinationCaller: []byte{2, 3, 4},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil recipient",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         nil,
				MessageBody:       []byte{2, 3, 4},
				DestinationCaller: []byte{2, 3, 4},
			},
			err: ErrMalformedField,
		},
		{
			name: "empty message body",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{},
				DestinationCaller: []byte{2, 3, 4},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil message body",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       nil,
				DestinationCaller: []byte{2, 3, 4},
			},
			err: ErrMalformedField,
		},
		{
			name: "empty destination caller",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{2, 3, 4},
				DestinationCaller: []byte{},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil destination caller",
			msg: MsgSendMessageWithCaller{
				From:              sample.AccAddress(),
				DestinationDomain: 123,
				Recipient:         []byte{2, 3, 4},
				MessageBody:       []byte{2, 3, 4},
				DestinationCaller: nil,
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
