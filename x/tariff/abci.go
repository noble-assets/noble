package tariff

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/noble/v5/x/tariff/keeper"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block
func BeginBlocker(ctx context.Context, k keeper.Keeper) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	k.AllocateTokens(sdkCtx)
}
