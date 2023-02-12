package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

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
		index := string(BlacklistedKey(elem.Address))
		if _, ok := blacklistedIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for blacklisted")
		}
		blacklistedIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in minters
	mintersIndexMap := make(map[string]struct{})

	for _, elem := range gs.MintersList {
		index := string(MintersKey(elem.Address))
		if _, ok := mintersIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for minters")
		}
		mintersIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in minterController
	minterControllerIndexMap := make(map[string]struct{})

	for _, elem := range gs.MinterControllerList {
		index := string(MinterControllerKey(elem.Minter))
		if _, ok := minterControllerIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for minterController")
		}
		minterControllerIndexMap[index] = struct{}{}
	}

	if gs.Owner == nil {
		return fmt.Errorf("tokenfactory owner account cannot be blank")
	}
	_, err := sdk.AccAddressFromBech32(gs.Owner.Address)
	if err != nil {
		return fmt.Errorf("invalid tokenfactory owner address")
	}

	if gs.Blacklister != nil {
		_, err := sdk.AccAddressFromBech32(gs.Blacklister.Address)
		if err != nil {
			return fmt.Errorf("invalid tokenfactory blacklister address")
		}
	}

	if gs.Pauser != nil {
		_, err := sdk.AccAddressFromBech32(gs.Pauser.Address)
		if err != nil {
			return fmt.Errorf("invalid tokenfactory pauser address")
		}
	}

	if gs.MintingDenom.Denom == "" {
		return fmt.Errorf("tokenfactory minting denom must be a registered in denom_metadata")
	}

	return gs.Params.Validate()
}
