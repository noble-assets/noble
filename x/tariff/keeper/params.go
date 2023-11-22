package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tariff/types"
=======
	"github.com/noble-assets/noble/v5/x/tariff/types"
>>>>>>> a4ad980 (chore: rename module path (#283))
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
