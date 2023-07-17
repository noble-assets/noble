package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDisableAttester = "disable_attester"

var _ sdk.Msg = &MsgDisableAttester{}

func NewMsgDisableAttester(from string, attester []byte) *MsgDisableAttester {
	return &MsgDisableAttester{
		From:     from,
		Attester: attester,
	}
}

func (msg *MsgDisableAttester) Route() string {
	return RouterKey
}

func (msg *MsgDisableAttester) Type() string {
	return TypeMsgDisableAttester
}

func (msg *MsgDisableAttester) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgDisableAttester) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDisableAttester) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	if msg.Attester == nil || len(msg.Attester) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Attester cannot be empty or nil")
	}
	return nil
}
