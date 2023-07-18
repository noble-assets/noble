package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAddPublicKey = "add_public_key"

var _ sdk.Msg = &MsgAddPublicKey{}

func NewMsgAddPublicKey(from string, publicKey []byte) *MsgAddPublicKey {
	return &MsgAddPublicKey{
		From:      from,
		PublicKey: publicKey,
	}
}

func (msg *MsgAddPublicKey) Route() string {
	return RouterKey
}

func (msg *MsgAddPublicKey) Type() string {
	return TypeMsgAddPublicKey
}

func (msg *MsgAddPublicKey) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgAddPublicKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAddPublicKey) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	if msg.PublicKey == nil || len(msg.PublicKey) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Public key cannot be empty or nil")
	}
	return nil
}
