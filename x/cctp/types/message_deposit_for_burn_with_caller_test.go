package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgDepositForBurnWithCaller_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDepositForBurnWithCaller
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgDepositForBurnWithCaller{
				From:              "invalid_address",
				Amount:            123,
				DestinationDomain: 123,
				MintRecipient:     []byte{1, 2, 3},
				BurnToken:         "utoken",
			},
			err: sdkerrors.ErrInvalidAddress,
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
