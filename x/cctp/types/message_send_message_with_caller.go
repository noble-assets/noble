package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSendMessageWithCaller = "send_message_with_caller"

var _ sdk.Msg = &MsgSendMessageWithCaller{}

func NewMsgSendMessageWithCaller(destinationDomain uint32, recipient []byte, messageBody []byte, destinationCaller []byte) *MsgSendMessageWithCaller {
	return &MsgSendMessageWithCaller{
		DestinationDomain: destinationDomain,
		Recipient:         recipient,
		MessageBody:       messageBody,
		DestinationCaller: destinationCaller,
	}
}

func (msg *MsgSendMessageWithCaller) Route() string {
	return RouterKey
}

func (msg *MsgSendMessageWithCaller) Type() string {
	return TypeMsgSendMessageWithCaller
}

func (msg *MsgSendMessageWithCaller) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgSendMessageWithCaller) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSendMessageWithCaller) ValidateBasic() error {
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
	if msg.DestinationCaller == nil || len(msg.DestinationCaller) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "DestinationCaller cannot be empty or nil")
	}

	return nil
}
