package keeper

import (
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBlacklisted set a specific blacklisted in the store from its index
func (k Keeper) SetBlacklisted(ctx sdk.Context, blacklisted types.Blacklisted) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklistedKeyPrefix))
	b := k.cdc.MustMarshal(&blacklisted)
	store.Set(types.BlacklistedKey(blacklisted.AddressBz), b)
}

// GetBlacklisted returns a blacklisted from its index
func (k Keeper) GetBlacklisted(ctx sdk.Context, addressBz []byte) (val types.Blacklisted, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklistedKeyPrefix))

	b := store.Get(types.BlacklistedKey(addressBz))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBlacklisted removes a blacklisted from the store
func (k Keeper) RemoveBlacklisted(ctx sdk.Context, addressBz []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklistedKeyPrefix))
	store.Delete(types.BlacklistedKey(addressBz))
}

// GetAllBlacklisted returns all blacklisted
func (k Keeper) GetAllBlacklisted(ctx sdk.Context) (list []types.Blacklisted) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BlacklistedKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Blacklisted
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
