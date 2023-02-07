package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgBlacklist_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgBlacklist
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgBlacklist{
				From:   "invalid_address",
				Pubkey: sample.PubKeyBytes(),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid pubkey",
			msg: MsgBlacklist{
				From:   sample.AccAddress(),
				Pubkey: []byte{},
			},
			err: ErrInvalidPubkey,
		},
		{
			name: "valid pubkey and from",
			msg: MsgBlacklist{
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
