package poa

import (
	"github.com/strangelove-ventures/noble/x/poa/keeper"
	"github.com/strangelove-ventures/noble/x/poa/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) error {
	// set the params
	k.SetParams(ctx, genState.Params)

	initVotes := len(genState.Votes) == 0

	for _, validator := range genState.Validators {
		k.SaveValidator(ctx, validator)

		if !initVotes {
			continue
		}

		// All genesis validators vote for each other
		for _, nestedVal := range genState.Validators {
			if nestedVal == validator {
				continue
			}
			k.SetVote(ctx, &types.Vote{
				VoterAddress:     validator.Address,
				CandidateAddress: nestedVal.Address,
				InFavor:          true,
			})
		}
	}

	err := k.CalculateValidatorVotes(ctx)
	if err != nil {
		return err
	}

	_, err = k.ApplyAndReturnValidatorSetUpdates(ctx)
	return err
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Validators = k.GetAllValidators(ctx)
	genesis.Votes = k.GetAllVotes(ctx)

	return genesis
}
