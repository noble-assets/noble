package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/x/cctp"
)

const TypeMsgLinkTokenPair = "link_token_pair"

var _ sdk.Msg = &MsgLinkTokenPair{}

func NewMsgLinkTokenPair(from string, remoteDomain uint32, remoteToken string, localToken string) *MsgLinkTokenPair {
	return &MsgLinkTokenPair{
		From:         from,
		RemoteDomain: remoteDomain,
		RemoteToken:  remoteToken,
		LocalToken:   localToken,
	}
}

func (msg *MsgLinkTokenPair) Route() string {
	return RouterKey
}

func (msg *MsgLinkTokenPair) Type() string {
	return TypeMsgLinkTokenPair
}

func (msg *MsgLinkTokenPair) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgLinkTokenPair) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgLinkTokenPair) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}
	if len(msg.RemoteToken) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Remote token is empty")
	}
	if len(msg.RemoteToken) > cctp.Bytes32Len {
		return sdkerrors.Wrapf(ErrMalformedField, "Remote token is over 32 bytes")
	}
	if len(msg.LocalToken) == 0 {
		return sdkerrors.Wrapf(ErrMalformedField, "Local token is empty")
	}
	return nil
}
