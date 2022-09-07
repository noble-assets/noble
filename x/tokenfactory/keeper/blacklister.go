package keeper

import (
	"noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBlacklister set blacklister in the store
func (k Keeper) SetBlacklister(ctx sdk.Context, blacklister types.Blacklister) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklisterKey))
	b := k.cdc.MustMarshal(&blacklister)
	store.Set([]byte{0}, b)
}

// GetBlacklister returns blacklister
func (k Keeper) GetBlacklister(ctx sdk.Context) (val types.Blacklister, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklisterKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBlacklister removes blacklister from the store
func (k Keeper) RemoveBlacklister(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklisterKey))
	store.Delete([]byte{0})
}
