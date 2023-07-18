package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgPauseSendingAndReceivingMessages = "pause_sending_and_receiving_messages"

var _ sdk.Msg = &MsgPauseSendingAndReceivingMessages{}

func NewMsgPauseSendingAndReceivingMessages(from string) *MsgPauseSendingAndReceivingMessages {
	return &MsgPauseSendingAndReceivingMessages{
		From: from,
	}
}

func (msg *MsgPauseSendingAndReceivingMessages) Route() string {
	return RouterKey
}

func (msg *MsgPauseSendingAndReceivingMessages) Type() string {
	return TypeMsgPauseSendingAndReceivingMessages
}

func (msg *MsgPauseSendingAndReceivingMessages) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgPauseSendingAndReceivingMessages) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPauseSendingAndReceivingMessages) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
