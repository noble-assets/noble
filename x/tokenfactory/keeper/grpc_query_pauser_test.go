package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/noble-assets/noble/v6/testutil/keeper"
	"github.com/noble-assets/noble/v6/testutil/nullify"
	"github.com/noble-assets/noble/v6/x/tokenfactory/types"
)

func TestPauserQuery(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	item := createTestPauser(keeper, ctx)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetPauserRequest
		response *types.QueryGetPauserResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetPauserRequest{},
			response: &types.QueryGetPauserResponse{Pauser: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Pauser(wctx, tc.request)
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
