package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/tariff/types"
)

// GetParams returns the params from state.
func (k *Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	bz := ctx.KVStore(k.storeKey).Get(types.ParamsKey)
	if bz == nil {
		panic("params not found in state")
	}

	k.cdc.MustUnmarshal(bz, &params)
	return
}

// SetParams stores the params in state.
func (k *Keeper) SetParams(ctx sdk.Context, params types.Params) {
	bz := k.cdc.MustMarshal(&params)
	ctx.KVStore(k.storeKey).Set(types.ParamsKey, bz)
}
