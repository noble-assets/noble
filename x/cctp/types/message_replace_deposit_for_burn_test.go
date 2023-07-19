package types

import (
	"github.com/strangelove-ventures/noble/testutil/sample"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgReplaceDepositForBurn_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgReplaceDepositForBurn
		err  error
	}{
		{
			name: "invalid from",
			msg: MsgReplaceDepositForBurn{
				From:                 "invalid_address",
				OriginalMessage:      []byte{1, 2, 3},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "nil original message",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      nil,
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty original message",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "nil original attestation",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{1, 2, 3},
				OriginalAttestation:  nil,
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty original attestation",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{1, 2, 3},
				OriginalAttestation:  []byte{},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "nil new destination caller",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{1, 2, 3},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: nil,
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty new destination caller",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{},
				NewMintRecipient:     []byte{1, 2, 3},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "nil new mint recipient",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{1, 2, 3},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     nil,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "empty new mint recipient",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "happy path",
			msg: MsgReplaceDepositForBurn{
				From:                 sample.AccAddress(),
				OriginalMessage:      []byte{1, 2, 3},
				OriginalAttestation:  []byte{1, 2, 3},
				NewDestinationCaller: []byte{1, 2, 3},
				NewMintRecipient:     []byte{1, 2, 3},
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
