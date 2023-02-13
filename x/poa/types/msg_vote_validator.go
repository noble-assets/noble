package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgVouchValidator{}

// Route should return the name of the module
func (msg *MsgVouchValidator) Route() string { return "poa" }

// Type should return the action
func (msg *MsgVouchValidator) Type() string { return "vouch_validator" }

// ValidateBasic runs stateless checks on the message
func (msg *MsgVouchValidator) ValidateBasic() error {
	if msg.GetCandidateAddress() == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "owner not specified")
	}
	if msg.GetVoucherAddress() == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "vouchr not specified")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgVouchValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.VoucherAddress)}
}

// GetSignBytes encodes the message for signing
func (msg *MsgVouchValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
