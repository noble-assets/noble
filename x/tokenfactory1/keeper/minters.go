package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/tokenfactory1/types"
)

// SetMinters set a specific minters in the store from its index
func (k Keeper) SetMinters(ctx sdk.Context, minters types.Minters) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintersKeyPrefix))
	b := k.cdc.MustMarshal(&minters)
	store.Set(types.MintersKey(
		minters.Address,
	), b)
}

// GetMinters returns a minters from its index
func (k Keeper) GetMinters(
	ctx sdk.Context,
	address string,

) (val types.Minters, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintersKeyPrefix))

	b := store.Get(types.MintersKey(
		address,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMinters removes a minters from the store
func (k Keeper) RemoveMinters(
	ctx sdk.Context,
	address string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintersKeyPrefix))
	store.Delete(types.MintersKey(
		address,
	))
}

// GetAllMinters returns all minters
func (k Keeper) GetAllMinters(ctx sdk.Context) (list []types.Minters) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintersKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Minters
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
