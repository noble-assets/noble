package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgConfigureMinter = "configure_minter"

var _ sdk.Msg = &MsgConfigureMinter{}

func NewMsgConfigureMinter(from string, address string, allowance sdk.Coin) *MsgConfigureMinter {
	return &MsgConfigureMinter{
		From:      from,
		Address:   address,
		Allowance: allowance,
	}
}

func (msg *MsgConfigureMinter) Route() string {
	return RouterKey
}

func (msg *MsgConfigureMinter) Type() string {
	return TypeMsgConfigureMinter
}

func (msg *MsgConfigureMinter) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgConfigureMinter) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgConfigureMinter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	_, err = sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid minter address (%s)", err)
	}

	if msg.Allowance.IsNil() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "allowance amount cannot be nil")
	}

	if msg.Allowance.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "allowance amount cannot be negative")
	}

	return nil
}
