package keeper

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPerMessageBurnLimit sets PerMessageBurnLimit in the store
func (k Keeper) SetPerMessageBurnLimit(ctx sdk.Context, amount types.PerMessageBurnLimit) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&amount)
	store.Set(types.KeyPrefix(types.PerMessageBurnLimitKey), b)
}

// GetPerMessageBurnLimit returns PerMessageBurnLimit
func (k Keeper) GetPerMessageBurnLimit(ctx sdk.Context) (val types.PerMessageBurnLimit, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.PerMessageBurnLimitKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
