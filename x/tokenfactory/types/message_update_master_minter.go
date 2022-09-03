package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateMasterMinter = "update_master_minter"

var _ sdk.Msg = &MsgUpdateMasterMinter{}

func NewMsgUpdateMasterMinter(from string, address string) *MsgUpdateMasterMinter {
	return &MsgUpdateMasterMinter{
		From:    from,
		Address: address,
	}
}

func (msg *MsgUpdateMasterMinter) Route() string {
	return RouterKey
}

func (msg *MsgUpdateMasterMinter) Type() string {
	return TypeMsgUpdateMasterMinter
}

func (msg *MsgUpdateMasterMinter) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUpdateMasterMinter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateMasterMinter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
