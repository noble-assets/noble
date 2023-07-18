package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnpauseSendingAndReceivingMessages = "unpause_sending_and_receiving_messages"

var _ sdk.Msg = &MsgUnpauseSendingAndReceivingMessages{}

func NewMsgUnpauseSendingAndReceivingMessages(from string) *MsgUnpauseSendingAndReceivingMessages {
	return &MsgUnpauseSendingAndReceivingMessages{
		From: from,
	}
}

func (msg *MsgUnpauseSendingAndReceivingMessages) Route() string {
	return RouterKey
}

func (msg *MsgUnpauseSendingAndReceivingMessages) Type() string {
	return TypeMsgUnpauseSendingAndReceivingMessages
}

func (msg *MsgUnpauseSendingAndReceivingMessages) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUnpauseSendingAndReceivingMessages) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnpauseSendingAndReceivingMessages) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
