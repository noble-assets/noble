package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/noble-assets/noble/v6/testutil/keeper"
	"github.com/noble-assets/noble/v6/testutil/nullify"
	"github.com/noble-assets/noble/v6/testutil/sample"
	"github.com/noble-assets/noble/v6/x/tokenfactory/keeper"
	"github.com/noble-assets/noble/v6/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

type blacklistedWrapper struct {
	address string
	bl      types.Blacklisted
}

func createNBlacklisted(keeper *keeper.Keeper, ctx sdk.Context, n int) []blacklistedWrapper {
	items := make([]blacklistedWrapper, n)
	for i := range items {
		acc := sample.TestAccount()
		items[i].address = acc.Address
		items[i].bl.AddressBz = acc.AddressBz

		keeper.SetBlacklisted(ctx, items[i].bl)
	}
	return items
}

func TestBlacklistedGet(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	items := createNBlacklisted(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBlacklisted(ctx,
			item.bl.AddressBz,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item.bl),
			nullify.Fill(&rst),
		)
	}
}

func TestBlacklistedRemove(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	items := createNBlacklisted(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBlacklisted(ctx,
			item.bl.AddressBz,
		)
		_, found := keeper.GetBlacklisted(ctx,
			item.bl.AddressBz,
		)
		require.False(t, found)
	}
}

func TestBlacklistedGetAll(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	items := createNBlacklisted(keeper, ctx, 10)
	blacklisted := make([]types.Blacklisted, len(items))
	for i, item := range items {
		blacklisted[i] = item.bl
	}
	require.ElementsMatch(t,
		nullify.Fill(blacklisted),
		nullify.Fill(keeper.GetAllBlacklisted(ctx)),
	)
}
