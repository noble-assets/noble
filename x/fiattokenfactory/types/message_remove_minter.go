package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRemoveMinter = "remove_minter"

var _ sdk.Msg = &MsgRemoveMinter{}

func NewMsgRemoveMinter(from, address string) *MsgRemoveMinter {
	return &MsgRemoveMinter{
		From:    from,
		Address: address,
	}
}

func (msg *MsgRemoveMinter) Route() string {
	return RouterKey
}

func (msg *MsgRemoveMinter) Type() string {
	return TypeMsgRemoveMinter
}

func (msg *MsgRemoveMinter) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgRemoveMinter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveMinter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid minter address (%s)", err)
	}
	return nil
}
