package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func TestBurningAndMintingPausedGet(t *testing.T) {

	keeper, ctx := keepertest.CctpKeeper(t)

	BurningAndMintingPaused := types.BurningAndMintingPaused{Paused: true}
	keeper.SetBurningAndMintingPaused(ctx, BurningAndMintingPaused)

	rst, found := keeper.GetBurningAndMintingPaused(ctx)
	require.True(t, found)
	require.Equal(t,
		BurningAndMintingPaused,
		nullify.Fill(&rst),
	)

	newBurningAndMintingPaused := types.BurningAndMintingPaused{Paused: false}

	keeper.SetBurningAndMintingPaused(ctx, newBurningAndMintingPaused)

	rst, found = keeper.GetBurningAndMintingPaused(ctx)
	require.True(t, found)
	require.Equal(t,
		newBurningAndMintingPaused,
		nullify.Fill(&rst),
	)
}
