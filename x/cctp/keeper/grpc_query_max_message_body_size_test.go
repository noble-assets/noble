package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func TestMaxMessageBodySizeQuery(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	MaxMessageBodySize := types.MaxMessageBodySize{Amount: 21}
	keeper.SetMaxMessageBodySize(ctx, MaxMessageBodySize)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMaxMessageBodySizeRequest
		response *types.QueryGetMaxMessageBodySizeResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetMaxMessageBodySizeRequest{},
			response: &types.QueryGetMaxMessageBodySizeResponse{Amount: MaxMessageBodySize},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.MaxMessageBodySize(wctx, tc.request)
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
