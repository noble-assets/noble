package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnpause = "unpause"

var _ sdk.Msg = &MsgUnpause{}

func NewMsgUnpause(from string) *MsgUnpause {
	return &MsgUnpause{
		From: from,
	}
}

func (msg *MsgUnpause) Route() string {
	return RouterKey
}

func (msg *MsgUnpause) Type() string {
	return TypeMsgUnpause
}

func (msg *MsgUnpause) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUnpause) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnpause) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
