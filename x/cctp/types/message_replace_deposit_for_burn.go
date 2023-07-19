package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgReplaceDepositForBurn = "replace_deposit_for_burn"

var _ sdk.Msg = &MsgReplaceDepositForBurn{}

func NewMsgReplaceDepositForBurn(originalMessage []byte, originalAttestation []byte, newDestinationCaller []byte, newMintRecipient []byte) *MsgReplaceDepositForBurn {
	return &MsgReplaceDepositForBurn{
		OriginalMessage:      originalMessage,
		OriginalAttestation:  originalAttestation,
		NewDestinationCaller: newDestinationCaller,
		NewMintRecipient:     newMintRecipient,
	}
}

func (msg *MsgReplaceDepositForBurn) Route() string {
	return RouterKey
}

func (msg *MsgReplaceDepositForBurn) Type() string {
	return TypeMsgReplaceDepositForBurn
}

func (msg *MsgReplaceDepositForBurn) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgReplaceDepositForBurn) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgReplaceDepositForBurn) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	if msg.OriginalAttestation == nil || len(msg.OriginalAttestation) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "OriginalAttestation cannot be empty or nil")
	}

	if msg.OriginalMessage == nil || len(msg.OriginalMessage) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "OriginalMessage cannot be empty or nil")
	}

	if msg.NewDestinationCaller == nil || len(msg.NewDestinationCaller) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "NewDestinationCaller cannot be empty or nil")
	}

	if msg.NewMintRecipient == nil || len(msg.NewMintRecipient) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "NewMintRecipient cannot be empty or nil")
	}

	return nil
}
