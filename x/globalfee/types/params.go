package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams creates a new Params object.
func NewParams(minimumGasPrices sdk.DecCoins, bypassMinFeeMsgTypes []string) Params {
	return Params{
		MinimumGasPrices:     minimumGasPrices,
		BypassMinFeeMsgTypes: bypassMinFeeMsgTypes,
	}
}

// DefaultParams creates a default Params object.
func DefaultParams() Params {
	return NewParams(sdk.NewDecCoins(), []string{
		"/ibc.core.client.v1.MsgUpdateClient",
		"/ibc.core.channel.v1.MsgRecvPacket",
		"/ibc.core.channel.v1.MsgAcknowledgement",
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/ibc.core.channel.v1.MsgTimeout",
		"/ibc.core.channel.v1.MsgTimeoutOnClose",
		"/cosmos.params.v1beta1.MsgUpdateParams",
		"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
		"/noble.fiattokenfactory.MsgUpdateMasterMinter",
		"/noble.fiattokenfactory.MsgUpdatePauser",
		"/noble.fiattokenfactory.MsgUpdateBlacklister",
		"/noble.fiattokenfactory.MsgUpdateOwner",
		"/noble.fiattokenfactory.MsgAcceptOwner",
		"/noble.fiattokenfactory.MsgConfigureMinter",
		"/noble.fiattokenfactory.MsgRemoveMinter",
		"/noble.fiattokenfactory.MsgMint",
		"/noble.fiattokenfactory.MsgBurn",
		"/noble.fiattokenfactory.MsgBlacklist",
		"/noble.fiattokenfactory.MsgUnblacklist",
		"/noble.fiattokenfactory.MsgPause",
		"/noble.fiattokenfactory.MsgUnpause",
		"/noble.fiattokenfactory.MsgConfigureMinterController",
		"/noble.fiattokenfactory.MsgRemoveMinterController",
		"/noble.tokenfactory.MsgUpdatePauser",
		"/noble.tokenfactory.MsgUpdateBlacklister",
		"/noble.tokenfactory.MsgUpdateOwner",
		"/noble.tokenfactory.MsgAcceptOwner",
		"/noble.tokenfactory.MsgConfigureMinter",
		"/noble.tokenfactory.MsgRemoveMinter",
		"/noble.tokenfactory.MsgMint",
		"/noble.tokenfactory.MsgBurn",
		"/noble.tokenfactory.MsgBlacklist",
		"/noble.tokenfactory.MsgUnblacklist",
		"/noble.tokenfactory.MsgPause",
		"/noble.tokenfactory.MsgUnpause",
		"/noble.tokenfactory.MsgConfigureMinterController",
		"/noble.tokenfactory.MsgRemoveMinterController",
	})
}

// Validate validates the provided params.
func (p *Params) Validate() error {
	if err := validateMinGasPrices(p.MinimumGasPrices); err != nil {
		return err
	}

	if err := validateBypassMinFeeMsgTypes(p.BypassMinFeeMsgTypes); err != nil {
		return err
	}

	return nil
}

//

func validateMinGasPrices(i interface{}) error {
	v, ok := i.(sdk.DecCoins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return v.Validate()
}

func validateBypassMinFeeMsgTypes(i interface{}) error {
	_, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	// todo: validate msg types are valid proto msg types?

	return nil
}
