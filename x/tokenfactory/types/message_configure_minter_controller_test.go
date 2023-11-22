package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/noble-assets/noble/v5/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgConfigureMinterController_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgConfigureMinterController
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgConfigureMinterController{
				From:       "invalid_address",
				Controller: sample.AccAddress(),
				Minter:     sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid controller",
			msg: MsgConfigureMinterController{
				From:       sample.AccAddress(),
				Controller: "invalid_address",
				Minter:     sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid minter",
			msg: MsgConfigureMinterController{
				From:       sample.AccAddress(),
				Controller: sample.AccAddress(),
				Minter:     "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address, minter, and controller",
			msg: MsgConfigureMinterController{
				From:       sample.AccAddress(),
				Controller: sample.AccAddress(),
				Minter:     sample.AccAddress(),
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
