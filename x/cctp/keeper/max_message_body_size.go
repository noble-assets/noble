package keeper

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMaxMessageBodySize sets MaxMessageBodySize in the store
func (k Keeper) SetMaxMessageBodySize(ctx sdk.Context, amount types.MaxMessageBodySize) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&amount)
	store.Set(types.KeyPrefix(types.MaxMessageBodySizeKey), b)
}

// GetMaxMessageBodySize returns MaxMessageBodySize
func (k Keeper) GetMaxMessageBodySize(ctx sdk.Context) (val types.MaxMessageBodySize, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.MaxMessageBodySizeKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
