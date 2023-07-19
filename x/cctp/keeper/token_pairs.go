package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetTokenPair sets a token pair in the store
func (k Keeper) SetTokenPair(ctx sdk.Context, tokenPair types.TokenPairs) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenPairKeyPrefix))
	b := k.cdc.MustMarshal(&tokenPair)
	store.Set(types.TokenPairKey(tokenPair.RemoteDomain, tokenPair.RemoteToken), b)
}

// GetTokenPair returns token pair
func (k Keeper) GetTokenPair(ctx sdk.Context, remoteDomain uint32, remoteToken string) (val types.TokenPairs, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenPairKeyPrefix))

	b := store.Get(types.TokenPairKey(remoteDomain, remoteToken))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteTokenPair removes a token pair from the store
func (k Keeper) DeleteTokenPair(
	ctx sdk.Context,
	remoteDomain uint32,
	remoteToken string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenPairKeyPrefix))
	store.Delete(types.TokenPairKey(remoteDomain, remoteToken))
}

// GetAllTokenPairs returns all token pairs
func (k Keeper) GetAllTokenPairs(ctx sdk.Context) (list []types.TokenPairs) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenPairKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.TokenPairs
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
