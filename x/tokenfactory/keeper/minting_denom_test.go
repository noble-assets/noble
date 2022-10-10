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

func createTestMintingDenom(keeper *keeper.Keeper, ctx sdk.Context) types.MintingDenom {
	item := types.MintingDenom{}
	keeper.SetMintingDenom(ctx, item)
	return item
}

func TestMintingDenomGet(t *testing.T) {
	keeper, ctx := keepertest.TokenfactoryKeeper(t)
	item := createTestMintingDenom(keeper, ctx)
	rst := keeper.GetMintingDenom(ctx)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
