package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/router/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMint sets a mint in the store
func (k Keeper) SetMint(ctx sdk.Context, key types.Mint) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MintPrefix(types.MintKeyPrefix))
	b := k.cdc.MustMarshal(&key)
	store.Set(types.LookupKey(key.SourceDomain, key.SourceDomainSender, key.Nonce), b)
}

// GetMint returns mint
func (k Keeper) GetMint(ctx sdk.Context, sourceDomain uint32, sourceDomainSender string, nonce uint64) (val types.Mint, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MintPrefix(types.MintKeyPrefix))

	b := store.Get(types.MintPrefix(string(types.LookupKey(sourceDomain, sourceDomainSender, nonce))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteMint removes a mint from the store
func (k Keeper) DeleteMint(ctx sdk.Context, sourceDomain uint32, sourceDomainSender string, nonce uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MintPrefix(types.MintKeyPrefix))
	store.Delete(types.MintPrefix(string(types.LookupKey(sourceDomain, sourceDomainSender, nonce))))
}

// GetAllMints returns all mints
func (k Keeper) GetAllMints(ctx sdk.Context) (list []types.Mint) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MintPrefix(types.MintKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Mint
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
