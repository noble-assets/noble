package tariff

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/v3/x/tariff/keeper"
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.AllocateTokens(ctx)
}
