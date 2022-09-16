package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

// SetMinterController set a specific minterController in the store from its index
func (k Keeper) SetMinterController(ctx sdk.Context, minterController types.MinterController) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))
	b := k.cdc.MustMarshal(&minterController)
	store.Set(types.MinterControllerKey(
		minterController.MinterAddress,
	), b)
}

// GetMinterController returns a minterController from its index
func (k Keeper) GetMinterController(
	ctx sdk.Context,
	minterAddress string,

) (val types.MinterController, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))

	b := store.Get(types.MinterControllerKey(
		minterAddress,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMinterController removes a minterController from the store
func (k Keeper) RemoveMinterController(
	ctx sdk.Context,
	minterAddress string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))
	store.Delete(types.MinterControllerKey(
		minterAddress,
	))
}

// GetAllMinterController returns all minterController
func (k Keeper) GetAllMinterController(ctx sdk.Context) (list []types.MinterController) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.MinterController
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
