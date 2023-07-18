package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnpauseBurningAndMinting = "pause_minting_and_burning"

var _ sdk.Msg = &MsgUnpauseBurningAndMinting{}

func NewMsgUnpauseBurningAndMinting(from string) *MsgUnpauseBurningAndMinting {
	return &MsgUnpauseBurningAndMinting{
		From: from,
	}
}

func (msg *MsgUnpauseBurningAndMinting) Route() string {
	return RouterKey
}

func (msg *MsgUnpauseBurningAndMinting) Type() string {
	return TypeMsgUnpauseBurningAndMinting
}

func (msg *MsgUnpauseBurningAndMinting) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUnpauseBurningAndMinting) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnpauseBurningAndMinting) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
