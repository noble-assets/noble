package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgRemovePublicKey_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRemovePublicKey
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgRemovePublicKey{
				From:      "invalid_address",
				PublicKey: []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty pubkey",
			msg: MsgRemovePublicKey{
				From:      sample.AccAddress(),
				PublicKey: []byte{},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil pubkey",
			msg: MsgRemovePublicKey{
				From:      sample.AccAddress(),
				PublicKey: nil,
			},
			err: ErrMalformedField,
		},
		{
			name: "valid address and key",
			msg: MsgRemovePublicKey{
				From:      sample.AccAddress(),
				PublicKey: []byte{1, 2, 3},
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
