package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddTokenMessenger = "add_token_messenger"

var _ sdk.Msg = &MsgAddTokenMessenger{}

func NewMsgAddTokenMessenger(from string, domainId uint32, address string) *MsgAddTokenMessenger {
	return &MsgAddTokenMessenger{
		From:     from,
		DomainId: domainId,
		Address:  address,
	}
}

func (msg *MsgAddTokenMessenger) Route() string {
	return RouterKey
}

func (msg *MsgAddTokenMessenger) Type() string {
	return TypeMsgAddTokenMessenger
}

func (msg *MsgAddTokenMessenger) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgAddTokenMessenger) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddTokenMessenger) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
