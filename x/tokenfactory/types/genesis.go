package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		BlacklistedList:      []Blacklisted{},
		Paused:               nil,
		MasterMinter:         nil,
		MintersList:          []Minters{},
		Pauser:               nil,
		Blacklister:          nil,
		Owner:                nil,
		MinterControllerList: []MinterController{},
		MintingDenom:         nil,
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in blacklisted
	blacklistedIndexMap := make(map[string]struct{})
	for _, elem := range gs.BlacklistedList {
		index := string(BlacklistedKey(elem.Pubkey))
		if _, ok := blacklistedIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for blacklisted")
		}
		blacklistedIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in minters and validate minter addr and allowance
	mintersIndexMap := make(map[string]struct{})
	for _, elem := range gs.MintersList {
		index := string(MintersKey(elem.Address))
		if _, ok := mintersIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for minters")
		}
		mintersIndexMap[index] = struct{}{}

		if _, err := sdk.AccAddressFromBech32(elem.Address); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid minter address (%s)", err)
		}

		if elem.Allowance.IsNil() || elem.Allowance.IsNegative() {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "minter allowance cannot be nil or negative")
		}
	}

	// Check for duplicated index in minterController and validate both controller and minter addresses
	minterControllerIndexMap := make(map[string]struct{})
	for _, elem := range gs.MinterControllerList {
		index := string(MinterControllerKey(elem.Controller))
		if _, ok := minterControllerIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for minterController")
		}
		minterControllerIndexMap[index] = struct{}{}

		if _, err := sdk.AccAddressFromBech32(elem.Minter); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "minter controller has invalid minter address (%s)", err)
		}

		if _, err := sdk.AccAddressFromBech32(elem.Controller); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "minter controller has invalid controller address (%s)", err)
		}
	}

	if gs.MasterMinter != nil {
		if _, err := sdk.AccAddressFromBech32(gs.MasterMinter.Address); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid master minter address (%s)", err)
		}
	}

	if gs.Pauser != nil {
		if _, err := sdk.AccAddressFromBech32(gs.Pauser.Address); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid pauser address (%s)", err)
		}
	}

	if gs.Blacklister != nil {
		if _, err := sdk.AccAddressFromBech32(gs.Blacklister.Address); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid black lister address (%s)", err)
		}
	}

	if gs.Owner != nil {
		if _, err := sdk.AccAddressFromBech32(gs.Owner.Address); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
		}
	}

	if gs.MintingDenom != nil {
		if gs.MintingDenom.Denom == "" {
			return fmt.Errorf("minting denom cannot be an empty string")
		}
	}

	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
