package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateAuthority_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateAuthority
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgUpdateAuthority{
				From:         "invalid_address",
				NewAuthority: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address",
			msg: MsgUpdateAuthority{
				From:         sample.AccAddress(),
				NewAuthority: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address and from",
			msg: MsgUpdateAuthority{
				From:         sample.AccAddress(),
				NewAuthority: sample.AccAddress(),
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
