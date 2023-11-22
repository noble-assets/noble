package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/noble-assets/noble/v4/testutil/keeper"
	"github.com/noble-assets/noble/v4/testutil/nullify"
	"github.com/noble-assets/noble/v4/x/tokenfactory/types"
)

func TestMintingDenomQuery(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	item := createTestMintingDenom(keeper, ctx)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMintingDenomRequest
		response *types.QueryGetMintingDenomResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetMintingDenomRequest{},
			response: &types.QueryGetMintingDenomResponse{MintingDenom: item},
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.MintingDenom(wctx, tc.request)
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
