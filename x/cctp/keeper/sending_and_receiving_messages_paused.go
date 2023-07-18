package keeper

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetSendingAndReceivingMessagesPaused set SendingAndReceivingMessagesPaused in the store
func (k Keeper) SetSendingAndReceivingMessagesPaused(ctx sdk.Context, paused types.SendingAndReceivingMessagesPaused) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&paused)
	store.Set(types.KeyPrefix(types.SendingAndReceivingMessagesPausedKey), b)
}

// GetSendingAndReceivingMessagesPaused returns SendingAndReceivingMessagesPaused
func (k Keeper) GetSendingAndReceivingMessagesPaused(ctx sdk.Context) (val types.SendingAndReceivingMessagesPaused, found bool) {
	store := ctx.KVStore(k.storeKey)

	b := store.Get(types.KeyPrefix(types.SendingAndReceivingMessagesPausedKey))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
