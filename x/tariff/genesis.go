package tariff

import (
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tariff/keeper"
	"github.com/strangelove-ventures/noble/v4/x/tariff/types"
=======
	"github.com/noble-assets/noble/v5/x/tariff/keeper"
	"github.com/noble-assets/noble/v5/x/tariff/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported GenesisState
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	return genesis
}
