package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNBlacklisted(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Blacklisted {
	items := make([]types.Blacklisted, n)
	for i := range items {
		items[i].Pubkey = []byte{byte(i)}

		keeper.SetBlacklisted(ctx, items[i])
	}
	return items
}

func TestBlacklistedGet(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	items := createNBlacklisted(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBlacklisted(ctx,
			item.Pubkey,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestBlacklistedRemove(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	items := createNBlacklisted(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBlacklisted(ctx,
			item.Pubkey,
		)
		_, found := keeper.GetBlacklisted(ctx,
			item.Pubkey,
		)
		require.False(t, found)
	}
}

func TestBlacklistedGetAll(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	items := createNBlacklisted(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBlacklisted(ctx)),
	)
}
