package types

import (
	"github.com/strangelove-ventures/noble/testutil/sample"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgUnpauseBurningAndMinting_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUnpauseBurningAndMinting
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUnpauseBurningAndMinting{
				From: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUnpauseBurningAndMinting{
				From: sample.AccAddress(),
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
