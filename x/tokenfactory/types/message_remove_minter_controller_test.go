package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/v5/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgRemoveMinterController_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRemoveMinterController
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgRemoveMinterController{
				From:       "invalid_address",
				Controller: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid controller",
			msg: MsgRemoveMinterController{
				From:       sample.AccAddress(),
				Controller: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid controller and from",
			msg: MsgRemoveMinterController{
				From:       sample.AccAddress(),
				Controller: sample.AccAddress(),
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
