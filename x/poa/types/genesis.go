package types

import (
	"encoding/base64"
	"fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Validators: []*Validator{},
		Params:     DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in blacklisted
	valAddressMap := make(map[string]struct{})
	valPubKeyMap := make(map[string]struct{})

	for _, val := range gs.Validators {
		address := base64.StdEncoding.EncodeToString(val.Address)
		if _, ok := valAddressMap[address]; ok {
			return fmt.Errorf("duplicated validator address: %s", address)
		}
		valAddressMap[address] = struct{}{}

		pubKey := base64.StdEncoding.EncodeToString(val.Pubkey.Value)
		if _, ok := valPubKeyMap[pubKey]; ok {
			return fmt.Errorf("duplicated validator pub key: %s", pubKey)
		}
		valPubKeyMap[pubKey] = struct{}{}
	}

	return gs.Params.Validate()
}
