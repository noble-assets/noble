package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgReplaceMessage = "replace_message"

var _ sdk.Msg = &MsgReplaceMessage{}

func NewMsgReplaceMessage(originalMessage []byte, originalAttestation []byte, newMessageBody []byte, newDestinationCaller []byte) *MsgReplaceMessage {
	return &MsgReplaceMessage{
		OriginalMessage:      originalMessage,
		OriginalAttestation:  originalAttestation,
		NewMessageBody:       newMessageBody,
		NewDestinationCaller: newDestinationCaller,
	}
}

func (msg *MsgReplaceMessage) Route() string {
	return RouterKey
}

func (msg *MsgReplaceMessage) Type() string {
	return TypeMsgReplaceMessage
}

func (msg *MsgReplaceMessage) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgReplaceMessage) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgReplaceMessage) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	if msg.OriginalMessage == nil || len(msg.OriginalMessage) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "original message must not be empty or nil: (%s)", err)
	}
	if msg.OriginalAttestation == nil || len(msg.OriginalAttestation) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "original attestation must not be empty or nil: (%s)", err)
	}
	if msg.NewMessageBody == nil || len(msg.NewMessageBody) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "new message body must not be empty or nil: (%s)", err)
	}
	return nil
}
