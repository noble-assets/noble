package keeper

import (
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

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
