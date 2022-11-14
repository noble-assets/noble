package keeper

import (
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetAdmin set admin in the store
func (k Keeper) SetAdmin(ctx sdk.Context, admin types.Admin) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&admin)
	store.Set(types.KeyPrefix(types.AdminKey), b)
}

// GetAdmin returns admin
func (k Keeper) GetAdmin(ctx sdk.Context) (val types.Admin, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.AdminKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
