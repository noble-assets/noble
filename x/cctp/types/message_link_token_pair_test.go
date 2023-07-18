package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgLinkTokenPair_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgLinkTokenPair
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgLinkTokenPair{
				From:         "invalid_address",
				RemoteDomain: 1,
				RemoteToken:  "0x12345",
				LocalToken:   "uusdc",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty remote token",
			msg: MsgLinkTokenPair{
				From:         sample.AccAddress(),
				RemoteDomain: 1,
				RemoteToken:  "",
				LocalToken:   "uusdc",
			},
			err: ErrMalformedField,
		},
		{
			name: "remote token too big",
			msg: MsgLinkTokenPair{
				From:         sample.AccAddress(),
				RemoteDomain: 1,
				RemoteToken:  "12345678901234567890123456789012345",
				LocalToken:   "uusdc",
			},
			err: ErrMalformedField,
		},
		{
			name: "empty local token",
			msg: MsgLinkTokenPair{
				From:         sample.AccAddress(),
				RemoteDomain: 1,
				RemoteToken:  "1234",
				LocalToken:   "",
			},
			err: ErrMalformedField,
		},
		{
			name: "valid address and fields",
			msg: MsgLinkTokenPair{
				From:         sample.AccAddress(),
				RemoteDomain: 1,
				RemoteToken:  "12345",
				LocalToken:   "uusdc",
			},
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
