package types

import (
	"fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		InFlightPackets: []InFlightPacket{},
		Mints:           []Mint{},
		IbcForwards:     []StoreIBCForwardMetadata{},
		Params:          DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	// Check for duplicated index in InFlightPackets
	inFlightPacketsIndexMap := make(map[string]struct{})
	for _, elem := range gs.InFlightPackets {
		index := string(LookupKey(elem.SourceDomain, elem.SourceDomainSender, elem.Nonce))
		if _, ok := inFlightPacketsIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for InFlightPackets")
		}
		inFlightPacketsIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in mints
	mintsIndexMap := make(map[string]struct{})
	for _, elem := range gs.Mints {
		index := string(LookupKey(elem.SourceDomain, elem.SourceDomainSender, elem.Nonce))
		if _, ok := mintsIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for Mints")
		}
		mintsIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in ibcForwards
	ibcForwardsIndexMap := make(map[string]struct{})
	for _, elem := range gs.IbcForwards {
		index := string(LookupKey(elem.SourceDomain, elem.SourceDomainSender, elem.Nonce))
		if _, ok := ibcForwardsIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for IBCForwards")
		}
		ibcForwardsIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
