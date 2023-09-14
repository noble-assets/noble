package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/strangelove-ventures/noble/v3/testutil/sample"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/strangelove-ventures/noble/v3/testutil/keeper"
	"github.com/strangelove-ventures/noble/v3/testutil/nullify"
	"github.com/strangelove-ventures/noble/v3/x/fiattokenfactory/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestBlacklistedQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBlacklisted(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetBlacklistedRequest
		response *types.QueryGetBlacklistedResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetBlacklistedRequest{
				Address: msgs[0].address,
			},
			response: &types.QueryGetBlacklistedResponse{Blacklisted: msgs[0].bl},
		},
		{
			desc: "Second",
			request: &types.QueryGetBlacklistedRequest{
				Address: msgs[1].address,
			},
			response: &types.QueryGetBlacklistedResponse{Blacklisted: msgs[1].bl},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetBlacklistedRequest{
				Address: sample.AccAddress(),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Blacklisted(wctx, tc.request)
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

func TestBlacklistedQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBlacklisted(keeper, ctx, 5)
	blacklisted := make([]types.Blacklisted, len(msgs))
	for i, msg := range msgs {
		blacklisted[i] = msg.bl
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllBlacklistedRequest {
		return &types.QueryAllBlacklistedRequest{
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
		for i := 0; i < len(blacklisted); i += step {
			resp, err := keeper.BlacklistedAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Blacklisted), step)
			require.Subset(t,
				nullify.Fill(blacklisted),
				nullify.Fill(resp.Blacklisted),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(blacklisted); i += step {
			resp, err := keeper.BlacklistedAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Blacklisted), step)
			require.Subset(t,
				nullify.Fill(blacklisted),
				nullify.Fill(resp.Blacklisted),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.BlacklistedAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(blacklisted), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(blacklisted),
			nullify.Fill(resp.Blacklisted),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.BlacklistedAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
