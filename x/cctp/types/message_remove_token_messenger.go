package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRemoveTokenMessenger = "remove_token_messenger"

var _ sdk.Msg = &MsgRemoveTokenMessenger{}

func NewMsgRemoveTokenMessenger(from string, domainId uint32) *MsgRemoveTokenMessenger {
	return &MsgRemoveTokenMessenger{
		From:     from,
		DomainId: domainId,
	}
}

func (msg *MsgRemoveTokenMessenger) Route() string {
	return RouterKey
}

func (msg *MsgRemoveTokenMessenger) Type() string {
	return TypeMsgRemoveTokenMessenger
}

func (msg *MsgRemoveTokenMessenger) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgRemoveTokenMessenger) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemoveTokenMessenger) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
