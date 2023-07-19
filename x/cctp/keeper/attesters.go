package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetAttester sets a public key in the store
func (k Keeper) SetAttester(ctx sdk.Context, key types.Attester) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AttesterKeyPrefix))
	b := k.cdc.MustMarshal(&key)
	store.Set(types.KeyPrefix(string(types.PublicKeyKey([]byte(key.Attester)))), b)
}

// GetAttester returns public key
func (k Keeper) GetAttester(ctx sdk.Context, key string) (val types.Attester, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AttesterKeyPrefix))

	b := store.Get(types.KeyPrefix(string(types.PublicKeyKey([]byte(key)))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteAttester removes a public key from the store
func (k Keeper) DeleteAttester(
	ctx sdk.Context,
	key string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AttesterKeyPrefix))
	store.Delete(types.PublicKeyKey(
		[]byte(key),
	))
}

// GetAllAttesters returns all public keys
func (k Keeper) GetAllAttesters(ctx sdk.Context) (list []types.Attester) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AttesterKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Attester
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
