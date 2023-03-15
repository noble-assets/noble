package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgConfigureMinter_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgConfigureMinter
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgConfigureMinter{
				From:    "invalid_address",
				Address: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address",
			msg: MsgConfigureMinter{
				From:    sample.AccAddress(),
				Address: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address and from",
			msg: MsgConfigureMinter{
				From:      sample.AccAddress(),
				Address:   sample.AccAddress(),
				Allowance: sdk.NewCoin("test", sdk.NewInt(1)),
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
