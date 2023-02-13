package types

import (
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	valAddressMap := make(map[string]struct{})
	valPubKeyMap := make(map[string]struct{})

	for _, val := range gs.Validators {
		// Check for duplicated validator address
		address := sdk.ValAddress(val.Address).String()
		if _, ok := valAddressMap[address]; ok {
			return fmt.Errorf("duplicated validator address: %s", address)
		}
		valAddressMap[address] = struct{}{}

		// Check for duplicated pub key
		pubKey := base64.StdEncoding.EncodeToString(val.Pubkey.Value)
		if _, ok := valPubKeyMap[pubKey]; ok {
			return fmt.Errorf("duplicated validator pub key: %s", pubKey)
		}
		valPubKeyMap[pubKey] = struct{}{}
	}

	votesMap := make(map[string]struct{})

	for _, vote := range gs.Votes {
		// Check for duplicated votes
		voter := sdk.ValAddress(vote.VoterAddress).String()
		candidate := sdk.ValAddress(vote.CandidateAddress).String()
		voterCandidateKey := voter + candidate
		if _, ok := votesMap[voterCandidateKey]; ok {
			return fmt.Errorf("duplicated vote from voter: %s for candidate: %s", voter, candidate)
		}
		valAddressMap[voterCandidateKey] = struct{}{}
	}

	return gs.Params.Validate()
}
