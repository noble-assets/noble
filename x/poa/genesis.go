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

	for _, validator := range genState.Validators {
		k.SetValidator(ctx, validator)
	}

	_, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	return err
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.Validators = k.GetAllValidators(ctx)

	return genesis
}
