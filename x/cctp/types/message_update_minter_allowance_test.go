package types

import (
	"testing"

	"github.com/strangelove-ventures/noble/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateMinterAllowance_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateMinterAllowance
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateMinterAllowance{
				From:   "invalid_address",
				Denom:  "asdf",
				Amount: 8000,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty denom",
			msg: MsgUpdateMinterAllowance{
				From:   sample.AccAddress(),
				Denom:  "",
				Amount: 8000,
			},
			err: ErrMalformedField,
		},
		{
			name: "happy path",
			msg: MsgUpdateMinterAllowance{
				From:   sample.AccAddress(),
				Denom:  "asdf",
				Amount: 8000,
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
