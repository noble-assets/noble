package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
<<<<<<< HEAD:x/fiattokenfactory/types/message_update_master_minter_test.go
	"github.com/strangelove-ventures/noble/testutil/sample"
=======
	"github.com/noble-assets/noble/v5/testutil/sample"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/types/message_update_master_minter_test.go
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateMasterMinter_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateMasterMinter
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgUpdateMasterMinter{
				From:    "invalid_address",
				Address: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address",
			msg: MsgUpdateMasterMinter{
				From:    sample.AccAddress(),
				Address: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address and from",
			msg: MsgUpdateMasterMinter{
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
