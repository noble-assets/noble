package tariff

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tariff/keeper"
=======
	"github.com/noble-assets/noble/v5/x/tariff/keeper"
>>>>>>> a4ad980 (chore: rename module path (#283))
)

// BeginBlocker sets the proposer for determining distribution during endblock
// and distribute rewards for the previous block
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.AllocateTokens(ctx)
}
