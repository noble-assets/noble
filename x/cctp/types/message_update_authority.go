package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateAuthority = "update_authority"

var _ sdk.Msg = &MsgUpdateAuthority{}

func NewMsgUpdateAuthority(from string, newAuthority string) *MsgUpdateAuthority {
	return &MsgUpdateAuthority{
		From:         from,
		NewAuthority: newAuthority,
	}
}

func (msg *MsgUpdateAuthority) Route() string {
	return RouterKey
}

func (msg *MsgUpdateAuthority) Type() string {
	return TypeMsgUpdateAuthority
}

func (msg *MsgUpdateAuthority) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdateAuthority) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateAuthority) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.NewAuthority)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid new authority address (%s)", err)
	}
	return nil
}
