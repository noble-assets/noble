package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

func createTestAdmin(keeper *keeper.Keeper, ctx sdk.Context) types.Admin {
	item := types.Admin{}
	keeper.SetAdmin(ctx, item)
	return item
}

func TestAdminGet(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	item := createTestAdmin(keeper, ctx)
	rst, found := keeper.GetAdmin(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
