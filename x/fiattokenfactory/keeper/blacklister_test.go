package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

<<<<<<< HEAD:x/fiattokenfactory/keeper/blacklister_test.go
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	keepertest "github.com/noble-assets/noble/v5/testutil/keeper"
	"github.com/noble-assets/noble/v5/testutil/nullify"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/keeper"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/blacklister_test.go
)

func createTestBlacklister(keeper *keeper.Keeper, ctx sdk.Context) types.Blacklister {
	item := types.Blacklister{}
	keeper.SetBlacklister(ctx, item)
	return item
}

func TestBlacklisterGet(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	item := createTestBlacklister(keeper, ctx)
	rst, found := keeper.GetBlacklister(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
