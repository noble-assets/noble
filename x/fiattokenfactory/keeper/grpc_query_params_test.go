package keeper_test

import (
	"testing"

	testkeeper "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := testkeeper.FiatTokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
