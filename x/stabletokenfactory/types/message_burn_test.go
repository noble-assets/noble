package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/testutil/sample"
=======
	"github.com/noble-assets/noble/v5/testutil/sample"
>>>>>>> a4ad980 (chore: rename module path (#283))
	"github.com/stretchr/testify/require"
)

func TestMsgBurn_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgBurn
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgBurn{
				From: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgBurn{
				From:   sample.AccAddress(),
				Amount: sdk.NewCoin("test", sdk.NewInt(1)),
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