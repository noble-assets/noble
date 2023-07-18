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

func TestPublicKeyQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPublicKeys(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetPublicKeyRequest
		response *types.QueryGetPublicKeyResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetPublicKeyRequest{
				Key: msgs[0].publicKey.Key,
			},
			response: &types.QueryGetPublicKeyResponse{PublicKey: msgs[0].publicKey},
		},
		{
			desc: "Second",
			request: &types.QueryGetPublicKeyRequest{
				Key: msgs[1].publicKey.Key,
			},
			response: &types.QueryGetPublicKeyResponse{PublicKey: msgs[1].publicKey},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetPublicKeyRequest{
				Key: "nothing",
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.PublicKey(wctx, tc.request)
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
	msgs := createNPublicKeys(keeper, ctx, 5)
	PublicKey := make([]types.PublicKeys, len(msgs))
	for i, msg := range msgs {
		PublicKey[i] = msg.publicKey
	}

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPublicKeysRequest {
		return &types.QueryAllPublicKeysRequest{
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
		for i := 0; i < len(PublicKey); i += step {
			resp, err := keeper.PublicKeys(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PublicKeys), step)
			require.Subset(t,
				nullify.Fill(PublicKey),
				nullify.Fill(resp.PublicKeys),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(PublicKey); i += step {
			resp, err := keeper.PublicKeys(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PublicKeys), step)
			require.Subset(t,
				nullify.Fill(PublicKey),
				nullify.Fill(resp.PublicKeys),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.PublicKeys(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(PublicKey), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(PublicKey),
			nullify.Fill(resp.PublicKeys),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.PublicKeys(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
