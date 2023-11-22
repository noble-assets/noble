package keeper

import (
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetOwner set owner in the store
func (k Keeper) SetOwner(ctx sdk.Context, owner types.Owner) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&owner)
	store.Set(types.KeyPrefix(types.OwnerKey), b)
}

// GetOwner returns owner
func (k Keeper) GetOwner(ctx sdk.Context) (val types.Owner, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.OwnerKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// SetPendingOwner set pending owner in the store
func (k Keeper) SetPendingOwner(ctx sdk.Context, owner types.Owner) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&owner)
	store.Set(types.KeyPrefix(types.PendingOwnerKey), b)
}

// DeletePendingOwner deletes the pending owner in the store
func (k Keeper) DeletePendingOwner(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.KeyPrefix(types.PendingOwnerKey))
}

// GetPendingOwner returns pending owner
func (k Keeper) GetPendingOwner(ctx sdk.Context) (val types.Owner, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.PendingOwnerKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
