package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

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
)

func createTestPauser(keeper *keeper.Keeper, ctx sdk.Context) types.Pauser {
	item := types.Pauser{}
	keeper.SetPauser(ctx, item)
	return item
}

func TestPauserGet(t *testing.T) {
	keeper, ctx := keepertest.StableTokenFactoryKeeper(t)
	item := createTestPauser(keeper, ctx)
	rst, found := keeper.GetPauser(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}