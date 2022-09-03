package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateBlacklister = "update_blacklister"

var _ sdk.Msg = &MsgUpdateBlacklister{}

func NewMsgUpdateBlacklister(from string, address string) *MsgUpdateBlacklister {
	return &MsgUpdateBlacklister{
		From:    from,
		Address: address,
	}
}

func (msg *MsgUpdateBlacklister) Route() string {
	return RouterKey
}

func (msg *MsgUpdateBlacklister) Type() string {
	return TypeMsgUpdateBlacklister
}

func (msg *MsgUpdateBlacklister) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdateBlacklister) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateBlacklister) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
