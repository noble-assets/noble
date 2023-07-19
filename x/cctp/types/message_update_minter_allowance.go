package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateMinterAllowance = "update_max_message_body_size"

var _ sdk.Msg = &MsgUpdateMinterAllowance{}

func NewMsgUpdateMinterAllowance(from string, denom string, amount uint64) *MsgUpdateMinterAllowance {
	return &MsgUpdateMinterAllowance{
		From:   from,
		Denom:  denom,
		Amount: amount,
	}
}

func (msg *MsgUpdateMinterAllowance) Route() string {
	return RouterKey
}

func (msg *MsgUpdateMinterAllowance) Type() string {
	return TypeMsgUpdateMinterAllowance
}

func (msg *MsgUpdateMinterAllowance) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdateMinterAllowance) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateMinterAllowance) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	return nil
}
