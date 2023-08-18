package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

var (
	_ legacytx.LegacyMsg = &MsgUpdateParams{}
	_ sdk.Msg            = &MsgUpdateParams{}
)

func (msg *MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshal(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	// NOTE: This can be removed when upgrading to Cosmos SDK Eden.
	authority, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{authority}
}

func (msg *MsgUpdateParams) Route() string {
	return ModuleName
}

func (msg *MsgUpdateParams) Type() string {
	return "noble/x/globalfee/MsgUpdateParams"
}

func (msg *MsgUpdateParams) ValidateBasic() error {
	// NOTE: This can be removed when upgrading to Cosmos SDK Eden.
	// https://docs.cosmos.network/v0.50/basics/tx-lifecycle#validatebasic-deprecated
	return nil
}
