package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDepositForBurn = "deposit_for_burn"

var _ sdk.Msg = &MsgDepositForBurn{}

func NewMsgDepositForBurn(amount uint32, destinationDomain uint32, mintRecipient []byte, burnToken string) *MsgDepositForBurn {
	return &MsgDepositForBurn{
		Amount:            amount,
		DestinationDomain: destinationDomain,
		MintRecipient:     mintRecipient,
		BurnToken:         burnToken,
	}
}

func (msg *MsgDepositForBurn) Route() string {
	return RouterKey
}

func (msg *MsgDepositForBurn) Type() string {
	return TypeMsgDepositForBurn
}

func (msg *MsgDepositForBurn) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgDepositForBurn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDepositForBurn) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
