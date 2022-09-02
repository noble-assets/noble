package keeper_test

import (
	"testing"

	testkeeper "noble/testutil/keeper"
	"noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParamsQuery(t *testing.T) {
	ctx, tk, _ := testkeeper.NewTestSetup(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	tk.TokenfactoryKeeper.SetParams(ctx, params)

	response, err := tk.TokenfactoryKeeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}
