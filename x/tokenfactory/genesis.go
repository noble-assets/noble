package tokenfactory

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/keeper"
	"noble/x/tokenfactory/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the blacklisted
	for _, elem := range genState.BlacklistedList {
		k.SetBlacklisted(ctx, elem)
	}
	// Set if defined
	if genState.Paused != nil {
		k.SetPaused(ctx, *genState.Paused)
	}
	// Set if defined
	if genState.MasterMinter != nil {
		k.SetMasterMinter(ctx, *genState.MasterMinter)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.BlacklistedList = k.GetAllBlacklisted(ctx)
	// Get all paused
	paused, found := k.GetPaused(ctx)
	if found {
		genesis.Paused = &paused
	}
	// Get all masterMinter
	masterMinter, found := k.GetMasterMinter(ctx)
	if found {
		genesis.MasterMinter = &masterMinter
	}
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
