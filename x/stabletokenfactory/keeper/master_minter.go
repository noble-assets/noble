package keeper

import (
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMasterMinter set masterMinter in the store
func (k Keeper) SetMasterMinter(ctx sdk.Context, masterMinter types.MasterMinter) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&masterMinter)
	store.Set(types.KeyPrefix(types.MasterMinterKey), b)
}

// GetMasterMinter returns masterMinter
func (k Keeper) GetMasterMinter(ctx sdk.Context) (val types.MasterMinter, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.MasterMinterKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}