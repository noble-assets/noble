package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

var _ legacytx.LegacyMsg = &MsgRegisterAccount{}

func (msg *MsgRegisterAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errors.New("invalid signer")
	}

	if !channeltypes.IsValidChannelID(msg.Channel) {
		return errors.New("invalid channel")
	}

	return nil
}

func (msg *MsgRegisterAccount) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{signer}
}

func (msg *MsgRegisterAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgRegisterAccount) Route() string {
	return ModuleName
}

func (msg *MsgRegisterAccount) Type() string {
	return "noble/forwarding/RegisterAccount"
}

//

var _ legacytx.LegacyMsg = &MsgClearAccount{}

func (msg *MsgClearAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errors.New("invalid signer")
	}

	_, err = sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return errors.New("invalid address")
	}

	return nil
}

func (msg *MsgClearAccount) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{signer}
}

func (msg *MsgClearAccount) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgClearAccount) Route() string {
	return ModuleName
}

func (msg *MsgClearAccount) Type() string {
	return "noble/forwarding/ClearAccount"
}
