package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

// SetMasterMinter set masterMinter in the store
func (k Keeper) SetMasterMinter(ctx sdk.Context, masterMinter types.MasterMinter) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MasterMinterKey))
	b := k.cdc.MustMarshal(&masterMinter)
	store.Set([]byte{0}, b)
}

// GetMasterMinter returns masterMinter
func (k Keeper) GetMasterMinter(ctx sdk.Context) (val types.MasterMinter, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MasterMinterKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMasterMinter removes masterMinter from the store
func (k Keeper) RemoveMasterMinter(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MasterMinterKey))
	store.Delete([]byte{0})
}
