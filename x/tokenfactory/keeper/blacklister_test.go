package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/v5/testutil/keeper"
	"github.com/strangelove-ventures/noble/v5/testutil/nullify"
	"github.com/strangelove-ventures/noble/v5/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/v5/x/tokenfactory/types"
)

func createTestBlacklister(keeper *keeper.Keeper, ctx sdk.Context) types.Blacklister {
	item := types.Blacklister{}
	keeper.SetBlacklister(ctx, item)
	return item
}

func TestBlacklisterGet(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	item := createTestBlacklister(keeper, ctx)
	rst, found := keeper.GetBlacklister(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
