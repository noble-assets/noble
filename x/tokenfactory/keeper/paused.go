package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

// SetPaused set paused in the store
func (k Keeper) SetPaused(ctx sdk.Context, paused types.Paused) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PausedKey))
	b := k.cdc.MustMarshal(&paused)
	store.Set([]byte{0}, b)
}

// GetPaused returns paused
func (k Keeper) GetPaused(ctx sdk.Context) (val types.Paused, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PausedKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePaused removes paused from the store
func (k Keeper) RemovePaused(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PausedKey))
	store.Delete([]byte{0})
}
