package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdatePerMessageBurnLimit = "update_per_message_burn_limit"

var _ sdk.Msg = &MsgUpdatePerMessageBurnLimit{}

func NewMsgUpdatePerMessageBurnLimit(from string, size uint64) *MsgUpdatePerMessageBurnLimit {
	return &MsgUpdatePerMessageBurnLimit{
		From:                from,
		PerMessageBurnLimit: size,
	}
}

func (msg *MsgUpdatePerMessageBurnLimit) Route() string {
	return RouterKey
}

func (msg *MsgUpdatePerMessageBurnLimit) Type() string {
	return TypeMsgUpdatePerMessageBurnLimit
}

func (msg *MsgUpdatePerMessageBurnLimit) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdatePerMessageBurnLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdatePerMessageBurnLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	return nil
}
