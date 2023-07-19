package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/router/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetIBCForward sets a IBCForward in the store
func (k Keeper) SetIBCForward(ctx sdk.Context, forward types.StoreIBCForwardMetadata) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.IBCForwardPrefix(types.IBCForwardKeyPrefix))
	b := k.cdc.MustMarshal(&forward)
	store.Set(types.LookupKey(forward.SourceDomain, forward.SourceDomainSender, forward.Metadata.Nonce), b)
}

// GetIBCForward returns IBCForward
func (k Keeper) GetIBCForward(ctx sdk.Context, sourceDomain uint32, sourceDomainSender string, nonce uint64) (val types.StoreIBCForwardMetadata, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.IBCForwardPrefix(types.IBCForwardKeyPrefix))

	b := store.Get(types.LookupKey(sourceDomain, sourceDomainSender, nonce))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteIBCForward removes a IBCForward from the store
func (k Keeper) DeleteIBCForward(ctx sdk.Context, sourceDomain uint32, sourceDomainSender string, nonce uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.IBCForwardPrefix(types.IBCForwardKeyPrefix))
	store.Delete(types.LookupKey(sourceDomain, sourceDomainSender, nonce))
}

// GetAllIBCForwards returns all IBCForwards
func (k Keeper) GetAllIBCForwards(ctx sdk.Context) (list []types.StoreIBCForwardMetadata) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.IBCForwardPrefix(types.IBCForwardKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StoreIBCForwardMetadata
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
