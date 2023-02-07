package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnblacklist = "unblacklist"

var _ sdk.Msg = &MsgUnblacklist{}

func NewMsgUnblacklist(from string, pubkey []byte) *MsgUnblacklist {
	return &MsgUnblacklist{
		From:   from,
		Pubkey: pubkey,
	}
}

func (msg *MsgUnblacklist) Route() string {
	return RouterKey
}

func (msg *MsgUnblacklist) Type() string {
	return TypeMsgUnblacklist
}

func (msg *MsgUnblacklist) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUnblacklist) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnblacklist) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	if len(msg.Pubkey) <= 0 {
		return sdkerrors.Wrap(ErrInvalidPubkey, "pubkey bytes cannot be less than or equal to 0")
	}
	return nil
}
