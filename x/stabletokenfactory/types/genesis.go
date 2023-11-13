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
		Params:               DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in blacklisted
	blacklistedIndexMap := make(map[string]struct{})
	for _, elem := range gs.BlacklistedList {
		index := string(BlacklistedKey(elem.AddressBz))
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

	var addresses []sdk.AccAddress

	if gs.Owner != nil {
		owner, err := sdk.AccAddressFromBech32(gs.Owner.Address)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid owner address (%s)", err)
		}
		addresses = append(addresses, owner)
	}

	if gs.MasterMinter != nil {
		masterMinter, err := sdk.AccAddressFromBech32(gs.MasterMinter.Address)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid master minter address (%s)", err)
		}
		addresses = append(addresses, masterMinter)
	}

	if gs.Pauser != nil {
		pauser, err := sdk.AccAddressFromBech32(gs.Pauser.Address)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid pauser address (%s)", err)
		}
		addresses = append(addresses, pauser)
	}

	if gs.Blacklister != nil {
		blacklister, err := sdk.AccAddressFromBech32(gs.Blacklister.Address)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid black lister address (%s)", err)
		}
		addresses = append(addresses, blacklister)
	}

	if err := validatePrivileges(addresses); err != nil {
		return err
	}

	if gs.MintingDenom != nil && gs.MintingDenom.Denom == "" {
		return fmt.Errorf("minting denom cannot be an empty string")
	}

	return gs.Params.Validate()
}

// validatePrivileges ensures that the same address is not being assigned to more than one privileged role.
func validatePrivileges(addresses []sdk.AccAddress) error {
	for i, current := range addresses {
		for j, target := range addresses {
			if i == j {
				continue
			}

			if current.String() == target.String() {
				return sdkerrors.Wrapf(ErrAlreadyPrivileged, "%s", current)
			}
		}
	}

	return nil
}
