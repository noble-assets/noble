package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// verify interface at compile time
var (
	_ sdk.Msg = &MsgRegisterAccount{}
	_ sdk.Msg = &MsgClearAccount{}
)

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
