package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "noble/testutil/keeper"
	"noble/testutil/nullify"
	"noble/x/tokenfactory/keeper"
	"noble/x/tokenfactory/types"
)

func createTestMasterMinter(keeper *keeper.Keeper, ctx sdk.Context) types.MasterMinter {
	item := types.MasterMinter{}
	keeper.SetMasterMinter(ctx, item)
	return item
}

func TestMasterMinterGet(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	item := createTestMasterMinter(keeper, ctx)
	rst, found := keeper.GetMasterMinter(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
