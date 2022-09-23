package keeper

import (
	"noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMinterController set a specific minterController in the store from its index
func (k Keeper) SetMinterController(ctx sdk.Context, minterController types.MinterController) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))
	b := k.cdc.MustMarshal(&minterController)
	store.Set(types.MinterControllerKey(
		minterController.Controller,
	), b)
}

// GetMinterController returns a minterController from its index
func (k Keeper) GetMinterController(
	ctx sdk.Context,
	address string,

) (val types.MinterController, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))

	b := store.Get(types.MinterControllerKey(
		address,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMinterController removes a minterController from the store
func (k Keeper) DeleteMinterController(
	ctx sdk.Context,
	address string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinterControllerKeyPrefix))
	store.Delete(types.MinterControllerKey(
		address,
	))
}

// GetAllMinterController returns all minterController
func (k Keeper) GetAllMinterControllers(ctx sdk.Context) (list []types.MinterController) {
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
