package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func TestPerMessageBurnLimitQuery(t *testing.T) {

	keeper, ctx := keepertest.CctpKeeper(t)

	PerMessageBurnLimit := types.PerMessageBurnLimit{Amount: 21}
	keeper.SetPerMessageBurnLimit(ctx, PerMessageBurnLimit)

	rst, found := keeper.GetPerMessageBurnLimit(ctx)
	require.True(t, found)
	require.Equal(t,
		PerMessageBurnLimit,
		nullify.Fill(&rst),
	)

	newPerMessageBurnLimit := types.PerMessageBurnLimit{Amount: 22}

	keeper.SetPerMessageBurnLimit(ctx, newPerMessageBurnLimit)

	rst, found = keeper.GetPerMessageBurnLimit(ctx)
	require.True(t, found)
	require.Equal(t,
		newPerMessageBurnLimit,
		nullify.Fill(&rst),
	)
}
