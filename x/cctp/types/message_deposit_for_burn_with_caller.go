package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDepositForBurnWithCaller = "deposit_for_burn_with_caller"

var _ sdk.Msg = &MsgDepositForBurnWithCaller{}

func NewMsgDepositForBurnWithCaller(amount uint32, destinationDomain uint32, mintRecipient []byte, burnToken string, destinationCaller []byte) *MsgDepositForBurnWithCaller {
	return &MsgDepositForBurnWithCaller{
		Amount:            amount,
		DestinationDomain: destinationDomain,
		MintRecipient:     mintRecipient,
		BurnToken:         burnToken,
		DestinationCaller: destinationCaller,
	}
}

func (msg *MsgDepositForBurnWithCaller) Route() string {
	return RouterKey
}

func (msg *MsgDepositForBurnWithCaller) Type() string {
	return TypeMsgDepositForBurnWithCaller
}

func (msg *MsgDepositForBurnWithCaller) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgDepositForBurnWithCaller) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDepositForBurnWithCaller) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	return nil
}
