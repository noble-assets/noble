package globalfee

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/v3/x/globalfee/types"
)

var _ types.QueryServer = &GrpcQuerier{}

// ParamSource is a read only subset of paramtypes.Subspace
type ParamSource interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Has(ctx sdk.Context, key []byte) bool
}

type GrpcQuerier struct {
	paramSource ParamSource
}

func NewGrpcQuerier(paramSource ParamSource) GrpcQuerier {
	return GrpcQuerier{paramSource: paramSource}
}

// Params returns the total set of global fee parameters.
func (g GrpcQuerier) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.Params, error) {
	var (
		minGasPrices         sdk.DecCoins
		bypassMinFeeMsgTypes []string
	)
	ctx := sdk.UnwrapSDKContext(stdCtx)
	if g.paramSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		g.paramSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &minGasPrices)
	}
	if g.paramSource.Has(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes) {
		g.paramSource.Get(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes, &bypassMinFeeMsgTypes)
	}
	return &types.Params{
		MinimumGasPrices:     minGasPrices,
		BypassMinFeeMsgTypes: bypassMinFeeMsgTypes,
	}, nil
}
