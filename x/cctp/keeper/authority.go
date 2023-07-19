package keeper

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetAuthority set authority in the store
func (k Keeper) SetAuthority(ctx sdk.Context, authority types.Authority) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&authority)
	store.Set(types.KeyPrefix(types.AuthorityKey), b)
}

// GetAuthority returns authority
func (k Keeper) GetAuthority(ctx sdk.Context) (val types.Authority, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.AuthorityKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
