package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/testutil/sample"
=======
	"github.com/noble-assets/noble/v5/testutil/sample"
>>>>>>> a4ad980 (chore: rename module path (#283))
	"github.com/stretchr/testify/require"
)

func TestMsgUnblacklist_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUnblacklist
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgUnblacklist{
				From:    "invalid_address",
				Address: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid block address",
			msg: MsgUnblacklist{
				From:    sample.AccAddress(),
				Address: "",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid block address and from",
			msg: MsgUnblacklist{
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
