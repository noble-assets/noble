package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSendMessage = "send_message"

var _ sdk.Msg = &MsgSendMessage{}

func NewMsgSendMessage(destinationDomain uint32, recipient []byte, messageBody []byte) *MsgSendMessage {
	return &MsgSendMessage{
		DestinationDomain: destinationDomain,
		Recipient:         recipient,
		MessageBody:       messageBody,
	}
}

func (msg *MsgSendMessage) Route() string {
	return RouterKey
}

func (msg *MsgSendMessage) Type() string {
	return TypeMsgSendMessage
}

func (msg *MsgSendMessage) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgSendMessage) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSendMessage) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	if msg.Recipient == nil || len(msg.Recipient) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Recipient cannot be empty or nil")
	}
	if msg.MessageBody == nil || len(msg.MessageBody) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Message body cannot be empty or nil")
	}
	return nil
}
