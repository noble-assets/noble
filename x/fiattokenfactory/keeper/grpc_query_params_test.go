package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD:x/fiattokenfactory/keeper/grpc_query_params_test.go
	testkeeper "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	testkeeper "github.com/noble-assets/noble/v5/testutil/keeper"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/grpc_query_params_test.go
	"github.com/stretchr/testify/require"
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
