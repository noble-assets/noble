package tokenfactory

import (
	"noble/x/tokenfactory/keeper"
	"noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// Set all the minters
	for _, elem := range genState.MintersList {
		k.SetMinters(ctx, elem)
	}
	// Set if defined
	if genState.Pauser != nil {
		k.SetPauser(ctx, *genState.Pauser)
	}
	// Set if defined
	if genState.Blacklister != nil {
		k.SetBlacklister(ctx, *genState.Blacklister)
	}
	// Set if defined
	if genState.Owner != nil {
		k.SetOwner(ctx, *genState.Owner)
	}
	// Set if defined
	if genState.Admin != nil {
		k.SetAdmin(ctx, *genState.Admin)
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
	paused := k.GetPaused(ctx)
	genesis.Paused = &paused

	// Get all masterMinter
	masterMinter, found := k.GetMasterMinter(ctx)
	if found {
		genesis.MasterMinter = &masterMinter
	}
	genesis.MintersList = k.GetAllMinters(ctx)
	// Get all pauser
	pauser, found := k.GetPauser(ctx)
	if found {
		genesis.Pauser = &pauser
	}
	// Get all blacklister
	blacklister, found := k.GetBlacklister(ctx)
	if found {
		genesis.Blacklister = &blacklister
	}
	// Get all owner
	owner, found := k.GetOwner(ctx)
	if found {
		genesis.Owner = &owner
	}
	// Get all admin
	admin, found := k.GetAdmin(ctx)
	if found {
		genesis.Admin = &admin
	}
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
