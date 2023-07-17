package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetAttester sets an attester in the store
func (k Keeper) SetAttester(ctx sdk.Context, attester types.Attester) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&attester)
	store.Set(types.KeyPrefix(types.AttesterKeyPrefix), b)
}

// GetAttester returns attester
func (k Keeper) GetAttester(ctx sdk.Context, attester string) (val types.Attester, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AttesterKeyPrefix))

	b := store.Get(types.KeyPrefix(string(types.AttesterKey([]byte(attester)))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteAttester removes a attester from the store
func (k Keeper) DeleteAttester(
	ctx sdk.Context,
	attester string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AttesterKeyPrefix))
	store.Delete(types.AttesterKey(
		[]byte(attester),
	))
}

// GetAllAttesters returns all attesters
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
