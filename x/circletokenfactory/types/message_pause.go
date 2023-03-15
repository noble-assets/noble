package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgPause = "pause"

var _ sdk.Msg = &MsgPause{}

func NewMsgPause(from string) *MsgPause {
	return &MsgPause{
		From: from,
	}
}

func (msg *MsgPause) Route() string {
	return RouterKey
}

func (msg *MsgPause) Type() string {
	return TypeMsgPause
}

func (msg *MsgPause) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgPause) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgPause) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
