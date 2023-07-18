package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgRemovePublicKey = "remove_public_key"

var _ sdk.Msg = &MsgRemovePublicKey{}

func NewMsgRemovePublicKey(from string, publicKey []byte) *MsgRemovePublicKey {
	return &MsgRemovePublicKey{
		From:      from,
		PublicKey: publicKey,
	}
}

func (msg *MsgRemovePublicKey) Route() string {
	return RouterKey
}

func (msg *MsgRemovePublicKey) Type() string {
	return TypeMsgRemovePublicKey
}

func (msg *MsgRemovePublicKey) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgRemovePublicKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRemovePublicKey) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	if msg.PublicKey == nil || len(msg.PublicKey) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Public key cannot be empty or nil")
	}
	return nil
}
