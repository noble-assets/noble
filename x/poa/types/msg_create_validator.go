package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ sdk.Msg = &MsgCreateValidator{}

// // MarshalJSON implements the json.Marshaler interface to provide custom JSON
// // serialization of the MsgCreateValidator type.
// // We define a custom marshaler here to allow for msg to be used in the genesis file
// func (msg MsgCreateValidator) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(msgCreateValidatorJSON{
// 		Name:        msg.Name,
// 		Address:     msg.Address,
// 		PubKey:      sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, msg.PubKey),
// 		Description: msg.Description,
// 		Owner:       msg.Owner,
// 	})
// }

// // UnmarshalJSON implements the json.Unmarshaler interface to provide custom
// // JSON deserialization of the MsgCreateValidatorPOA type.
// func (msg *MsgCreateValidator) UnmarshalJSON(bz []byte) error {
// 	var msgCreateValJSON msgCreateValidatorJSON
// 	if err := json.Unmarshal(bz, &msgCreateValJSON); err != nil {
// 		return err
// 	}

// 	msg.Name = msgCreateValJSON.Name
// 	msg.Address = msgCreateValJSON.Address
// 	var err error
// 	msg.PubKey, err = sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, msgCreateValJSON.PubKey)
// 	if err != nil {
// 		return err
// 	}
// 	msg.Description = msgCreateValJSON.Description
// 	msg.Owner = msgCreateValJSON.Owner

// 	return nil
// }

// Route should return the name of the module
func (msg *MsgCreateValidator) Route() string { return "poa" }

// Type should return the action
func (msg *MsgCreateValidator) Type() string { return "create_validator" }

// ValidateBasic runs stateless checks on the message
func (msg *MsgCreateValidator) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, fmt.Sprintf("address must be bech32 with prefix: %s", sdk.Bech32PrefixAccAddr))
	}

	if msg.Description == (stakingtypes.Description{}) {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	return nil
}

// GetSigners defines whose signature is required
func (msg *MsgCreateValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(msg.Address)}
}

// GetSignBytes encodes the message for signing
func (msg *MsgCreateValidator) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}
