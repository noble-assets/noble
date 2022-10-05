package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const TypeMsgSoftwareUpgrade = "software_upgrade"

var _ sdk.Msg = &MsgSoftwareUpgrade{}

func NewMsgSoftwareUpgrade(from string, plan upgradetypes.Plan) *MsgSoftwareUpgrade {
	return &MsgSoftwareUpgrade{
		From: from,
		Plan: &plan,
	}
}

func (msg *MsgSoftwareUpgrade) Route() string {
	return RouterKey
}

func (msg *MsgSoftwareUpgrade) Type() string {
	return TypeMsgSoftwareUpgrade
}

func (msg *MsgSoftwareUpgrade) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgSoftwareUpgrade) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSoftwareUpgrade) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
