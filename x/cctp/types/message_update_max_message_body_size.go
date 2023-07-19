package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateMaxMessageBodySize = "update_max_message_body_size"

var _ sdk.Msg = &MsgUpdateMaxMessageBodySize{}

func NewMsgUpdateMaxMessageBodySize(from string, size uint32) *MsgUpdateMaxMessageBodySize {
	return &MsgUpdateMaxMessageBodySize{
		From:        from,
		MessageSize: size,
	}
}

func (msg *MsgUpdateMaxMessageBodySize) Route() string {
	return RouterKey
}

func (msg *MsgUpdateMaxMessageBodySize) Type() string {
	return TypeMsgUpdateMaxMessageBodySize
}

func (msg *MsgUpdateMaxMessageBodySize) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdateMaxMessageBodySize) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateMaxMessageBodySize) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	return nil
}
