package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetUsedNonce sets a nonce in the store
func (k Keeper) SetUsedNonce(ctx sdk.Context, key types.Nonce) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))
	store.Set(key.Bytes(), []byte{1})
}

// GetUsedNonce returns nonce
func (k Keeper) GetUsedNonce(ctx sdk.Context, key types.Nonce) (found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))

	return store.Get(key.Bytes()) != nil
}

// GetAllUsedNonces returns all UsedNonces
func (k Keeper) GetAllUsedNonces(ctx sdk.Context) (list []types.Nonce) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Nonce
		key := iterator.Key()
		if err := val.Unmarshal(key); err != nil {
			panic(err)
		}

		list = append(list, val)
	}

	return
}
