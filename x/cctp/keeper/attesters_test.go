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

type attesterWrapper struct {
	address  string
	attester types.Attester
}

func createNAttesters(keeper *keeper.Keeper, ctx sdk.Context, n int) []attesterWrapper {
	items := make([]attesterWrapper, n)
	for i := range items {
		items[i].address = sample.AccAddress()
		items[i].attester.Attester = "PublicKey" + strconv.Itoa(i)

		keeper.SetAttester(ctx, items[i].attester)
	}
	return items
}

func TestAttestersGet(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNAttesters(cctpKeeper, ctx, 10)
	for _, item := range items {
		rst, found := cctpKeeper.GetAttester(ctx,
			item.attester.Attester,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item.attester),
			nullify.Fill(&rst),
		)
	}
}

func TestAttestersRemove(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNAttesters(cctpKeeper, ctx, 10)
	for _, item := range items {
		cctpKeeper.DeleteAttester(ctx, item.address)
		_, found := cctpKeeper.GetAttester(ctx, item.address)
		require.False(t, found)
	}
}

func TestAttestersGetAll(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNAttesters(cctpKeeper, ctx, 10)
	denom := make([]types.Attester, len(items))
	for i, item := range items {
		denom[i] = item.attester
	}
	require.ElementsMatch(t,
		nullify.Fill(denom),
		nullify.Fill(cctpKeeper.GetAllAttesters(ctx)),
	)
}
