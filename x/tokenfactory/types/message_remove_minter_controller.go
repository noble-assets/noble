package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRemoveMinterController = "remove_minter_controller"

var _ sdk.Msg = &MsgRemoveMinterController{}

func NewMsgRemoveMinterController(from string, address string) *MsgRemoveMinterController {
	return &MsgRemoveMinterController{
		From:    from,
		Address: address,
	}
}

func (msg *MsgRemoveMinterController) Route() string {
	return RouterKey
}

func (msg *MsgRemoveMinterController) Type() string {
	return TypeMsgRemoveMinterController
}

func (msg *MsgRemoveMinterController) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgRemoveMinterController) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveMinterController) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
