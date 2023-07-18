package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func TestMaxMessageBodySizeGet(t *testing.T) {

	keeper, ctx := keepertest.CctpKeeper(t)

	MaxMessageBodySize := types.MaxMessageBodySize{Amount: 21}
	keeper.SetMaxMessageBodySize(ctx, MaxMessageBodySize)

	rst, found := keeper.GetMaxMessageBodySize(ctx)
	require.True(t, found)
	require.Equal(t,
		MaxMessageBodySize,
		nullify.Fill(&rst),
	)

	newMaxMessageBodySize := types.MaxMessageBodySize{Amount: 22}

	keeper.SetMaxMessageBodySize(ctx, newMaxMessageBodySize)

	rst, found = keeper.GetMaxMessageBodySize(ctx)
	require.True(t, found)
	require.Equal(t,
		newMaxMessageBodySize,
		nullify.Fill(&rst),
	)
}
