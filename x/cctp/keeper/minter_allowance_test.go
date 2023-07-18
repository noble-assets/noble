package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/testutil/sample"
	"github.com/strangelove-ventures/noble/x/cctp/keeper"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

type minterAllowanceWrapper struct {
	address         string
	minterAllowance types.MinterAllowances
}

func createNMinterAllowances(keeper *keeper.Keeper, ctx sdk.Context, n int) []minterAllowanceWrapper {
	items := make([]minterAllowanceWrapper, n)
	for i := range items {
		items[i].address = sample.AccAddress()
		items[i].minterAllowance.Denom = "denom" + strconv.Itoa(i)
		items[i].minterAllowance.Amount = uint64(i)

		keeper.SetMinterAllowance(ctx, items[i].minterAllowance)
	}
	return items
}

func TestMinterAllowancesGet(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNMinterAllowances(cctpKeeper, ctx, 10)
	for _, item := range items {
		rst, found := cctpKeeper.GetMinterAllowance(ctx,
			item.minterAllowance.Denom,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item.minterAllowance),
			nullify.Fill(&rst),
		)
	}
}

func TestMinterAllowancesRemove(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNMinterAllowances(cctpKeeper, ctx, 10)
	for _, item := range items {
		cctpKeeper.DeleteMinterAllowance(ctx, item.address)
		_, found := cctpKeeper.GetMinterAllowance(ctx, item.address)
		require.False(t, found)
	}
}

func TestMinterAllowancesGetAll(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNMinterAllowances(cctpKeeper, ctx, 10)
	denom := make([]types.MinterAllowances, len(items))
	for i, item := range items {
		denom[i] = item.minterAllowance
	}
	require.ElementsMatch(t,
		nullify.Fill(denom),
		nullify.Fill(cctpKeeper.GetAllMinterAllowances(ctx)),
	)
}
