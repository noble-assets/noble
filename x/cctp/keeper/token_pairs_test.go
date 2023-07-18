package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/strangelove-ventures/noble/x/cctp/keeper"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

type tokenPairWrapper struct {
	address   string
	tokenPair types.TokenPairs
}

func createNTokenPairs(keeper *keeper.Keeper, ctx sdk.Context, n int) []tokenPairWrapper {
	items := make([]tokenPairWrapper, n)
	for i := range items {
		items[i].address = sample.AccAddress()
		items[i].tokenPair.RemoteDomain = uint32(i)
		items[i].tokenPair.RemoteToken = strconv.Itoa(i)
		items[i].tokenPair.LocalToken = "token" + strconv.Itoa(i)

		keeper.SetTokenPair(ctx, items[i].tokenPair)
	}
	return items
}

func TestTokenPairsGet(t *testing.T) {

	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNTokenPairs(cctpKeeper, ctx, 10)
	for _, item := range items {
		rst, found := cctpKeeper.GetTokenPair(ctx,
			item.tokenPair.RemoteDomain,
			item.tokenPair.RemoteToken,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item.tokenPair),
			nullify.Fill(&rst),
		)
	}
}

func TestTokenPairsRemove(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNTokenPairs(cctpKeeper, ctx, 10)
	for _, item := range items {
		cctpKeeper.DeleteTokenPair(
			ctx,
			item.tokenPair.RemoteDomain,
			item.tokenPair.RemoteToken,
		)
		_, found := cctpKeeper.GetTokenPair(
			ctx,
			item.tokenPair.RemoteDomain,
			item.tokenPair.RemoteToken,
		)
		require.False(t, found)
	}
}

func TestTokenPairsGetAll(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNTokenPairs(cctpKeeper, ctx, 10)
	denom := make([]types.TokenPairs, len(items))
	for i, item := range items {
		denom[i] = item.tokenPair
	}
	require.ElementsMatch(t,
		nullify.Fill(denom),
		nullify.Fill(cctpKeeper.GetAllTokenPairs(ctx)),
	)
}
