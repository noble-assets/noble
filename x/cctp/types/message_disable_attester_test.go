package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgDisableAttester_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDisableAttester
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDisableAttester{
				From:     "invalid_address",
				Attester: []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty pubkey",
			msg: MsgDisableAttester{
				From:     sample.AccAddress(),
				Attester: []byte{},
			},
			err: ErrMalformedField,
		},
		{
			name: "nil pubkey",
			msg: MsgDisableAttester{
				From:     sample.AccAddress(),
				Attester: nil,
			},
			err: ErrMalformedField,
		},
		{
			name: "valid address and key",
			msg: MsgDisableAttester{
				From:     sample.AccAddress(),
				Attester: []byte{1, 2, 3},
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
