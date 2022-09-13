package keeper

import (
	"noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPauser set pauser in the store
func (k Keeper) SetPauser(ctx sdk.Context, pauser types.Pauser) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&pauser)
	store.Set(types.KeyPrefix(types.PauserKey), b)
}

// GetPauser returns pauser
func (k Keeper) GetPauser(ctx sdk.Context) (val types.Pauser, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.PauserKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePauser removes pauser from the store
func (k Keeper) RemovePauser(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PauserKey))
	store.Delete(types.KeyPrefix(types.PauserKey))
}
