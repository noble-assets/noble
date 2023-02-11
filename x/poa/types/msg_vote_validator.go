package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgVoteValidator{}

// Route should return the name of the module
func (msg *MsgVoteValidator) Route() string { return "poa" }

// Type should return the action
func (msg *MsgVoteValidator) Type() string { return "vote_validator" }

// ValidateBasic runs stateless checks on the message
func (msg *MsgVoteValidator) ValidateBasic() error {
	if msg.GetCandidateAddress() == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "owner not specified")
	}
	if msg.GetVoterAddress() == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "voter not specified")
	}
	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgVoteValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.VoterAddress)}
}

// GetSignBytes encodes the message for signing
func (msg *MsgVoteValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
