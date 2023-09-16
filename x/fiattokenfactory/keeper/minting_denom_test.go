package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/v4/testutil/keeper"
	"github.com/strangelove-ventures/noble/v4/testutil/nullify"
	"github.com/strangelove-ventures/noble/v4/x/fiattokenfactory/keeper"
	"github.com/strangelove-ventures/noble/v4/x/fiattokenfactory/types"
)

func createTestMintingDenom(keeper *keeper.Keeper, ctx sdk.Context) types.MintingDenom {
	item := types.MintingDenom{
		Denom: "abcd",
	}
	keeper.SetMintingDenom(ctx, item)
	return item
}

func TestMintingDenomGet(t *testing.T) {
	keeper, ctx := keepertest.FiatTokenfactoryKeeper(t)
	item := createTestMintingDenom(keeper, ctx)
	rst := keeper.GetMintingDenom(ctx)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}
