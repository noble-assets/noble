package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestTokenMessengerQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNTokenMessengers(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetTokenMessengerRequest
		response *types.QueryGetTokenMessengerResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetTokenMessengerRequest{
				DomainId: 1,
			},
			response: &types.QueryGetTokenMessengerResponse{TokenMessenger: msgs[0].tokenMessenger},
		},
		{
			desc: "Second",
			request: &types.QueryGetTokenMessengerRequest{
				DomainId: 2,
			},
			response: &types.QueryGetTokenMessengerResponse{TokenMessenger: msgs[1].tokenMessenger},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetTokenMessengerRequest{
				DomainId: 1111,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.TokenMessenger(wctx, tc.request)
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

func TestTokenMessengerQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNTokenMessengers(keeper, ctx, 5)
	TokenMessenger := make([]types.TokenMessenger, len(msgs))
	for i, msg := range msgs {
		TokenMessenger[i] = msg.attester
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllTokenMessengersRequest {
		return &types.QueryAllTokenMessengersRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(TokenMessenger); i += step {
			resp, err := keeper.TokenMessengers(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.TokenMessenger), step)
			require.Subset(t,
				nullify.Fill(TokenMessenger),
				nullify.Fill(resp.TokenMessengers),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(TokenMessenger); i += step {
			resp, err := keeper.TokenMessengers(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.TokenMessengers), step)
			require.Subset(t,
				nullify.Fill(TokenMessenger),
				nullify.Fill(resp.TokenMessengers),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.TokenMessengers(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(TokenMessenger), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(TokenMessenger),
			nullify.Fill(resp.TokenMessengers),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.TokenMessengers(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
