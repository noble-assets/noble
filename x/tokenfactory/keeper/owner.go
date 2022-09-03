package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

// SetOwner set owner in the store
func (k Keeper) SetOwner(ctx sdk.Context, owner types.Owner) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	b := k.cdc.MustMarshal(&owner)
	store.Set([]byte{0}, b)
}

// GetOwner returns owner
func (k Keeper) GetOwner(ctx sdk.Context) (val types.Owner, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveOwner removes owner from the store
func (k Keeper) RemoveOwner(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OwnerKey))
	store.Delete([]byte{0})
}
