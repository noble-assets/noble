package keeper

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBurningAndMintingPaused set BurningAndMintingPaused in the store
func (k Keeper) SetBurningAndMintingPaused(ctx sdk.Context, paused types.BurningAndMintingPaused) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&paused)
	store.Set(types.KeyPrefix(types.BurningAndMintingPausedKey), b)
}

// GetBurningAndMintingPaused returns BurningAndMintingPaused
func (k Keeper) GetBurningAndMintingPaused(ctx sdk.Context) (val types.BurningAndMintingPaused, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.BurningAndMintingPausedKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
