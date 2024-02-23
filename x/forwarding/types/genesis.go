package types

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (gen *GenesisState) Validate() error {
	for channel := range gen.NumOfAccounts {
		if !channeltypes.IsValidChannelID(channel) {
			return errors.New("invalid channel")
		}
	}

	for channel := range gen.NumOfForwards {
		if !channeltypes.IsValidChannelID(channel) {
			return errors.New("invalid channel")
		}
	}

	for channel, total := range gen.TotalForwarded {
		if !channeltypes.IsValidChannelID(channel) {
			return errors.New("invalid channel")
		}

		if _, err := sdk.ParseCoinsNormalized(total); err != nil {
			return errors.New("invalid coins")
		}
	}

	return nil
}
