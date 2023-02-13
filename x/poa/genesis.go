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

	initVouches := len(genState.Vouches) == 0

	for _, validator := range genState.Validators {
		k.SaveValidator(ctx, validator)

		if !initVouches {
			continue
		}

		// All genesis validators vouch for each other
		for _, nestedVal := range genState.Validators {
			if nestedVal == validator {
				continue
			}
			k.SetVouch(ctx, &types.Vouch{
				VoucherAddress:   validator.Address,
				CandidateAddress: nestedVal.Address,
				InFavor:          true,
			})
		}
	}

	err := k.CalculateValidatorVouches(ctx)
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
	genesis.Vouches = k.GetAllVouches(ctx)

	return genesis
}
