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

func TestAttesterQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNAttesters(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetAttesterRequest
		response *types.QueryGetAttesterResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetAttesterRequest{
				Attester: msgs[0].attester.Attester,
			},
			response: &types.QueryGetAttesterResponse{Attester: msgs[0].attester},
		},
		{
			desc: "Second",
			request: &types.QueryGetAttesterRequest{
				Attester: msgs[1].attester.Attester,
			},
			response: &types.QueryGetAttesterResponse{Attester: msgs[1].attester},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetAttesterRequest{
				Attester: "nothing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Attester(wctx, tc.request)
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

func TestPublicKeyQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNAttesters(keeper, ctx, 5)
	attesters := make([]types.Attester, len(msgs))
	for i, msg := range msgs {
		attesters[i] = msg.attester
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllAttestersRequest {
		return &types.QueryAllAttestersRequest{
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
		for i := 0; i < len(attesters); i += step {
			resp, err := keeper.Attesters(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Attester), step)
			require.Subset(t,
				nullify.Fill(attesters),
				nullify.Fill(resp.Attester),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(attesters); i += step {
			resp, err := keeper.Attesters(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Attester), step)
			require.Subset(t,
				nullify.Fill(attesters),
				nullify.Fill(resp.Attester),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.Attesters(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(attesters), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(attesters),
			nullify.Fill(resp.Attester),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.Attesters(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
