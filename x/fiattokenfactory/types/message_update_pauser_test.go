package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
<<<<<<< HEAD:x/fiattokenfactory/types/message_update_pauser_test.go
	"github.com/strangelove-ventures/noble/testutil/sample"
=======
	"github.com/noble-assets/noble/v5/testutil/sample"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/types/message_update_pauser_test.go
	"github.com/stretchr/testify/require"
)

func TestMsgUpdatePauser_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdatePauser
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgUpdatePauser{
				From:    "invalid_address",
				Address: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address",
			msg: MsgUpdatePauser{
				From:    sample.AccAddress(),
				Address: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address and from",
			msg: MsgUpdatePauser{
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
