package types

import (
	"testing"

	"github.com/strangelove-ventures/noble/testutil/sample"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateMaxMessageBodySize_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateMaxMessageBodySize
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateMaxMessageBodySize{
				From:        "invalid_address",
				MessageSize: 8000,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "happy path",
			msg: MsgUpdateMaxMessageBodySize{
				From:        sample.AccAddress(),
				MessageSize: 8000,
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
