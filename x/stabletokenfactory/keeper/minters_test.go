package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	keepertest "github.com/strangelove-ventures/noble/v4/testutil/keeper"
	"github.com/strangelove-ventures/noble/v4/testutil/nullify"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/keeper"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	keepertest "github.com/noble-assets/noble/v5/testutil/keeper"
	"github.com/noble-assets/noble/v5/testutil/nullify"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/keeper"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNMinters(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Minters {
	items := make([]types.Minters, n)
	for i := range items {
		items[i].Address = strconv.Itoa(i)

		keeper.SetMinters(ctx, items[i])
	}
	return items
}

func TestMintersGet(t *testing.T) {
	keeper, ctx := keepertest.StableTokenFactoryKeeper(t)
	items := createNMinters(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetMinters(ctx,
			item.Address,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}

func TestMintersRemove(t *testing.T) {
	keeper, ctx := keepertest.StableTokenFactoryKeeper(t)
	items := createNMinters(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveMinters(ctx,
			item.Address,
		)
		_, found := keeper.GetMinters(ctx,
			item.Address,
		)
		require.False(t, found)
	}
}

func TestMintersGetAll(t *testing.T) {
	keeper, ctx := keepertest.StableTokenFactoryKeeper(t)
	items := createNMinters(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllMinters(ctx)),
	)
}
