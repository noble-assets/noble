package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/strangelove-ventures/noble/v4/testutil/keeper"
	"github.com/strangelove-ventures/noble/v4/testutil/nullify"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
)

func TestOwnerQuery(t *testing.T) {
	keeper, ctx := keepertest.StableTokenFactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	owner := types.Owner{Address: "test"}
	keeper.SetOwner(ctx, owner)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetOwnerRequest
		response *types.QueryGetOwnerResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetOwnerRequest{},
			response: &types.QueryGetOwnerResponse{Owner: owner},
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
