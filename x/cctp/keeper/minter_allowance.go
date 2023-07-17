package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMinterAllowance sets a minter allowance in the store
func (k Keeper) SetMinterAllowance(ctx sdk.Context, allowance types.MinterAllowances) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&allowance)
	store.Set(types.KeyPrefix(types.MinterAllowanceKeyPrefix), b)
}

// GetMinterAllowance returns minter allowance
func (k Keeper) GetMinterAllowance(ctx sdk.Context, denom string) (val types.MinterAllowances, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterAllowanceKeyPrefix))

	b := store.Get(types.KeyPrefix(string(types.MinterAllowanceKey([]byte(denom)))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteMinterAllowance removes a minter allowance from the store
func (k Keeper) DeleteMinterAllowance(
	ctx sdk.Context,
	denom string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterAllowanceKeyPrefix))
	store.Delete(types.MinterAllowanceKey(
		[]byte(denom),
	))
}

// GetAllMinterAllowances returns all minter allowances
func (k Keeper) GetAllMinterAllowances(ctx sdk.Context) (list []types.MinterAllowances) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterAllowanceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.MinterAllowances
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
