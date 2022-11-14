package keeper

import (
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBlacklister set blacklister in the store
func (k Keeper) SetBlacklister(ctx sdk.Context, blacklister types.Blacklister) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&blacklister)
	store.Set(types.KeyPrefix(types.BlacklisterKey), b)
}

// GetBlacklister returns blacklister
func (k Keeper) GetBlacklister(ctx sdk.Context) (val types.Blacklister, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.BlacklisterKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBlacklister removes blacklister from the store
func (k Keeper) RemoveBlacklister(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyPrefix(types.BlacklisterKey))
}
