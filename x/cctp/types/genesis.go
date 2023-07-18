package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Authority:                         nil,
		PublicKeysList:                    []PublicKeys{},
		MinterAllowanceList:               []MinterAllowances{},
		PerMessageBurnLimit:               nil,
		BurningAndMintingPaused:           nil,
		SendingAndReceivingMessagesPaused: nil,
		MaxMessageBodySize:                nil,
		Nonce:                             nil,
		Params:                            DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Authority != nil {
		_, err := sdk.AccAddressFromBech32(gs.Authority.Address)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
		}
	}

	// Check for duplicated index in public keys
	publicKeysIndexMap := make(map[string]struct{})
	for _, elem := range gs.PublicKeysList {
		index := string(PublicKeyKey([]byte(elem.Key)))
		if _, ok := publicKeysIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for public keys")
		}
		publicKeysIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in minter allowance
	minterAllowancesIndexMap := make(map[string]struct{})
	for _, elem := range gs.MinterAllowanceList {
		index := string(MinterAllowanceKey([]byte(elem.Denom)))
		if _, ok := minterAllowancesIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for public keys")
		}
		minterAllowancesIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
