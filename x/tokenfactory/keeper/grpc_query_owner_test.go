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

func TestOwnerQuery(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	item := createTestOwner(keeper, ctx)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetOwnerRequest
		response *types.QueryGetOwnerResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetOwnerRequest{},
			response: &types.QueryGetOwnerResponse{Owner: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Owner(wctx, tc.request)
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
