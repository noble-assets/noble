package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgBlacklist = "blacklist"

var _ sdk.Msg = &MsgBlacklist{}

func NewMsgBlacklist(from string, address string) *MsgBlacklist {
	return &MsgBlacklist{
		From:    from,
		Address: address,
	}
}

func (msg *MsgBlacklist) Route() string {
	return RouterKey
}

func (msg *MsgBlacklist) Type() string {
	return TypeMsgBlacklist
}

func (msg *MsgBlacklist) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgBlacklist) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgBlacklist) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid blacklist address (%s)", err)
	}
	return nil
}
