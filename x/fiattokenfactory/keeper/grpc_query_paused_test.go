package keeper_test

import (
	"testing"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestPausedQuery(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	item := createTestPaused(keeper, ctx)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetPausedRequest
		response *types.QueryGetPausedResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetPausedRequest{},
			response: &types.QueryGetPausedResponse{Paused: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Paused(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}
