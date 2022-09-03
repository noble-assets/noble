package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

// SetPauser set pauser in the store
func (k Keeper) SetPauser(ctx sdk.Context, pauser types.Pauser) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PauserKey))
	b := k.cdc.MustMarshal(&pauser)
	store.Set([]byte{0}, b)
}

// GetPauser returns pauser
func (k Keeper) GetPauser(ctx sdk.Context) (val types.Pauser, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PauserKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePauser removes pauser from the store
func (k Keeper) RemovePauser(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PauserKey))
	store.Delete([]byte{0})
}
