package tariff

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/tariff/keeper"
	"github.com/strangelove-ventures/noble/x/tariff/types"
)

func InitGenesis(ctx sdk.Context, k *keeper.Keeper, gs types.GenesisState) {
	k.SetParams(ctx, gs.Params)
}

func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)
	return types.NewGenesisState(params)
}
