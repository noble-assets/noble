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
		AttesterList:                      []Attester{},
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
	var addresses []sdk.AccAddress

	if gs.Authority == nil {
		return fmt.Errorf("authority cannot be nil")
	}

	if gs.Authority != nil {
		owner, err := sdk.AccAddressFromBech32(gs.Authority.Address)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address (%s)", err)
		}
		addresses = append(addresses, owner)
	}

	if gs.AttesterList == nil {
		return fmt.Errorf("AttesterList cannot be nil")
	}

	// Check for duplicated index in attesters
	attesterIndexMap := make(map[string]struct{})
	for _, elem := range gs.AttesterList {
		index := string(AttesterKey([]byte(elem.Attester)))
		if _, ok := attesterIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for attesters")
		}
		attesterIndexMap[index] = struct{}{}
	}

	if gs.MinterAllowanceList == nil {
		return fmt.Errorf("minter allowance cannot be nil")
	}

	// Check for duplicated index in minter allowance
	minterAllowancesIndexMap := make(map[string]struct{})
	for _, elem := range gs.MinterAllowanceList {
		index := string(MinterAllowanceKey([]byte(elem.Denom)))
		if _, ok := minterAllowancesIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for attesters")
		}
		minterAllowancesIndexMap[index] = struct{}{}
	}

	if gs.PerMessageBurnLimit == nil || gs.PerMessageBurnLimit.Amount < 0 {
		return fmt.Errorf("per message burn limit cannot be nil or less than 0")
	}

	if gs.MaxMessageBodySize != nil && gs.MaxMessageBodySize.Amount < 0 {
		return fmt.Errorf("max message body size cannot be less than 0")
	}

	if gs.BurningAndMintingPaused == nil {
		return fmt.Errorf("BurningAndMintingPaused cannot be nil")
	}

	if gs.SendingAndReceivingMessagesPaused == nil {
		return fmt.Errorf("SendingAndReceivingMessagesPaused cannot be nil")
	}

	if gs.MaxMessageBodySize == nil {
		return fmt.Errorf("MaxMessageBodySize cannot be nil")
	}

	if gs.Nonce == nil {
		return fmt.Errorf("nonce cannot be nil")
	}

	return gs.Params.Validate()
}
