package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		BlacklistedList: []Blacklisted{},
		Paused:          nil,
		MasterMinter:    nil,
		MintersList:     []Minters{},
		Pauser:          nil,
		Blacklister:     nil,
		Owner:           nil,
		Admin:           nil,
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
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
