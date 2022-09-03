package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

// SetAdmin set admin in the store
func (k Keeper) SetAdmin(ctx sdk.Context, admin types.Admin) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AdminKey))
	b := k.cdc.MustMarshal(&admin)
	store.Set([]byte{0}, b)
}

// GetAdmin returns admin
func (k Keeper) GetAdmin(ctx sdk.Context) (val types.Admin, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AdminKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveAdmin removes admin from the store
func (k Keeper) RemoveAdmin(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AdminKey))
	store.Delete([]byte{0})
}
