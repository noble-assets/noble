package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
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
				From:   "invalid_address",
				Pubkey: sample.PubKeyBytes(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid pubkey",
			msg: MsgUnblacklist{
				From:   sample.AccAddress(),
				Pubkey: []byte{},
			},
			err: ErrInvalidPubkey,
		},
		{
			name: "valid pubkey and from",
			msg: MsgUnblacklist{
				From:   sample.AccAddress(),
				Pubkey: sample.PubKeyBytes(),
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
