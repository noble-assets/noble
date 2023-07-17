package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/x/cctp"
)

const TypeMsgUnlinkTokenPair = "unlink_token_pair"

var _ sdk.Msg = &MsgUnlinkTokenPair{}

func NewMsgUnlinkTokenPair(from string, remoteDomain uint32, remoteToken string, localToken string) *MsgUnlinkTokenPair {
	return &MsgUnlinkTokenPair{
		From:         from,
		RemoteDomain: remoteDomain,
		RemoteToken:  remoteToken,
		LocalToken:   localToken,
	}
}

func (msg *MsgUnlinkTokenPair) Route() string {
	return RouterKey
}

func (msg *MsgUnlinkTokenPair) Type() string {
	return TypeMsgUnlinkTokenPair
}

func (msg *MsgUnlinkTokenPair) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}

func (msg *MsgUnlinkTokenPair) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnlinkTokenPair) ValidateBasic() error {
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
