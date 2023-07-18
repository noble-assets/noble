package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetSignatureThreshold sets a SignatureThreshold in the store
func (k Keeper) SetSignatureThreshold(ctx sdk.Context, key types.SignatureThreshold) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SignatureThresholdKeyPrefix))
	b := k.cdc.MustMarshal(&key)
	store.Set(types.KeyPrefix(types.SignatureThresholdKeyPrefix), b)
}

// GetSignatureThreshold returns SignatureThreshold
func (k Keeper) GetSignatureThreshold(ctx sdk.Context) (val types.SignatureThreshold, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SignatureThresholdKeyPrefix))

	b := store.Get(types.KeyPrefix(types.SignatureThresholdKeyPrefix))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
