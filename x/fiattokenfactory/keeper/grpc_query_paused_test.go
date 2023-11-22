package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

<<<<<<< HEAD:x/fiattokenfactory/keeper/grpc_query_paused_test.go
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	keepertest "github.com/noble-assets/noble/v5/testutil/keeper"
	"github.com/noble-assets/noble/v5/testutil/nullify"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/grpc_query_paused_test.go
)

func TestPausedQuery(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
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
