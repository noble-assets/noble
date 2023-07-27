package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgConfigureMinterController = "configure_minter_controller"

var _ sdk.Msg = &MsgConfigureMinterController{}

func NewMsgConfigureMinterController(from, controller, minter string) *MsgConfigureMinterController {
	return &MsgConfigureMinterController{
		From:       from,
		Controller: controller,
		Minter:     minter,
	}
}

func (msg *MsgConfigureMinterController) Route() string {
	return RouterKey
}

func (msg *MsgConfigureMinterController) Type() string {
	return TypeMsgConfigureMinterController
}

func (msg *MsgConfigureMinterController) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgConfigureMinterController) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgConfigureMinterController) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.Controller)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid controller address (%s)", err)
	}
	_, err = sdk.AccAddressFromBech32(msg.Minter)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid minter address (%s)", err)
	}
	return nil
}
