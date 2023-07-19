package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetNonce sets a nonce in the store
func (k Keeper) SetNonce(ctx sdk.Context, key types.Nonce) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&key)
	store.Set(types.KeyPrefix(types.NonceKeyPrefix), b)
}

// GetNonce returns nonce
func (k Keeper) GetNonce(ctx sdk.Context) (val types.Nonce, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.NonceKeyPrefix))

	b := store.Get(types.KeyPrefix(types.NonceKeyPrefix))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
