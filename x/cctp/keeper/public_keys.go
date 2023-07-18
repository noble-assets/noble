package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPublicKey sets a public key in the store
func (k Keeper) SetPublicKey(ctx sdk.Context, key types.PublicKeys) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&key)
	store.Set(types.KeyPrefix(types.PublicKeyKeyPrefix), b)
}

// GetPublicKey returns public key
func (k Keeper) GetPublicKey(ctx sdk.Context, key string) (val types.PublicKeys, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PublicKeyKeyPrefix))

	b := store.Get(types.KeyPrefix(string(types.PublicKeyKey([]byte(key)))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeletePublicKey removes a public key from the store
func (k Keeper) DeletePublicKey(
	ctx sdk.Context,
	key string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PublicKeyKeyPrefix))
	store.Delete(types.PublicKeyKey(
		[]byte(key),
	))
}

// GetAllPublicKeys returns all public keys
func (k Keeper) GetAllPublicKeys(ctx sdk.Context) (list []types.PublicKeys) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PublicKeyKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PublicKeys
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
