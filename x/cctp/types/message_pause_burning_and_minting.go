package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgPauseBurningAndMinting = "pause_minting_and_burning"

var _ sdk.Msg = &MsgPauseBurningAndMinting{}

func NewMsgPauseBurningAndMinting(from string) *MsgPauseBurningAndMinting {
	return &MsgPauseBurningAndMinting{
		From: from,
	}
}

func (msg *MsgPauseBurningAndMinting) Route() string {
	return RouterKey
}

func (msg *MsgPauseBurningAndMinting) Type() string {
	return TypeMsgPauseBurningAndMinting
}

func (msg *MsgPauseBurningAndMinting) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgPauseBurningAndMinting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPauseBurningAndMinting) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
