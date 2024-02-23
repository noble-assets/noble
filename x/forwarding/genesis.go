package forwarding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/noble/v5/x/forwarding/keeper"
	"github.com/noble-assets/noble/v5/x/forwarding/types"
)

func InitGenesis(ctx sdk.Context, k *keeper.Keeper, genesis types.GenesisState) {
	for channel, count := range genesis.NumOfAccounts {
		k.SetNumOfAccounts(ctx, channel, count)
	}

	for channel, count := range genesis.NumOfForwards {
		k.SetNumOfForwards(ctx, channel, count)
	}

	for channel, rawTotal := range genesis.TotalForwarded {
		total, _ := sdk.ParseCoinsNormalized(rawTotal)
		k.SetTotalForwarded(ctx, channel, total)
	}
}

func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		NumOfAccounts:  k.GetAllNumOfAccounts(ctx),
		NumOfForwards:  k.GetAllNumOfForwards(ctx),
		TotalForwarded: k.GetAllTotalForwarded(ctx),
	}
}
