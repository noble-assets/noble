package types

import (
	"testing"

	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func TestMsgUpdateBlacklister_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateBlacklister
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgUpdateBlacklister{
				From:    "invalid_address",
				Address: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address",
			msg: MsgUpdateBlacklister{
				From:    sample.AccAddress(),
				Address: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address and from",
			msg: MsgUpdateBlacklister{
				From:    sample.AccAddress(),
				Address: sample.AccAddress(),
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
