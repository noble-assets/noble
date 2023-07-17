package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetUsedNonce sets a nonce in the store
func (k Keeper) SetUsedNonce(ctx sdk.Context, key types.Nonce) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&key)
	store.Set(types.KeyPrefix(types.UsedNonceKeyPrefix), b)
}

// GetUsedNonce returns nonce
func (k Keeper) GetUsedNonce(ctx sdk.Context, key uint64) (val types.Nonce, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))

	var byteKey []byte
	binary.BigEndian.PutUint64(byteKey, key)
	b := store.Get(types.KeyPrefix(string(types.UsedNonceKey(byteKey))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllUsedNonces returns all UsedNonces
func (k Keeper) GetAllUsedNonces(ctx sdk.Context) (list []types.Nonce) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Nonce
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
