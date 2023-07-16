package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnblacklist = "unblacklist"

var _ sdk.Msg = &MsgUnblacklist{}

func NewMsgUnblacklist(from, address string) *MsgUnblacklist {
	return &MsgUnblacklist{
		From:    from,
		Address: address,
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

	if len(msg.Address) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "address length cannot be less than or equal to 0")
	}
	return nil
}
