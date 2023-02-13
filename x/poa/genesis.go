package poa

import (
	"github.com/strangelove-ventures/noble/x/poa/keeper"
	"github.com/strangelove-ventures/noble/x/poa/types"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genState types.GenesisState) ([]abci.ValidatorUpdate, error) {
	// set the params
	k.SetParams(ctx, genState.Params)

	for _, validator := range genState.Validators {
		k.SaveValidator(ctx, validator)

		k.AfterValidatorCreated(ctx, validator.GetOperator())

		consAddr, err := validator.GetConsAddr()
		if err != nil {
			return nil, err
		}

		k.AfterValidatorBonded(ctx, consAddr, validator.GetOperator())
	}

	for _, vouch := range genState.Vouches {
		k.SetVouch(ctx, vouch)
	}

	err := k.CalculateValidatorVouches(ctx)
	if err != nil {
		return nil, err
	}

	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)

	k.Logger(ctx).Info("Returning genesis val updates", "count", len(updates), "updates", updates)

	return updates, err
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Validators = k.GetAllValidators(ctx)
	genesis.Vouches = k.GetAllVouches(ctx)

	return genesis
}
