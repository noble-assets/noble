package keeper

import (
<<<<<<< HEAD:x/fiattokenfactory/keeper/paused.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/paused.go

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPaused set paused in the store
func (k Keeper) SetPaused(ctx sdk.Context, paused types.Paused) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&paused)
	store.Set(types.KeyPrefix(types.PausedKey), b)
}

// GetPaused returns paused
func (k Keeper) GetPaused(ctx sdk.Context) (val types.Paused) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.PausedKey))
	if b == nil {
		panic("Paused state is not set")
	}

	k.cdc.MustUnmarshal(b, &val)
	return val
}
