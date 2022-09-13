package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "noble/testutil/keeper"
	"noble/testutil/nullify"
	"noble/x/tokenfactory/types"
)

func TestPausedQuery(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
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
