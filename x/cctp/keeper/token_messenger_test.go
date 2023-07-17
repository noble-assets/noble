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

type tokenMessengerWrapper struct {
	address        string
	tokenMessenger types.TokenMessenger
}

func createNTokenMessengers(keeper *keeper.Keeper, ctx sdk.Context, n int) []tokenMessengerWrapper {
	items := make([]tokenMessengerWrapper, n)
	for i := range items {
		items[i].address = sample.AccAddress()
		items[i].tokenMessenger.DomainId = uint32(i)

		keeper.SetTokenMessenger(ctx, items[i].tokenMessenger)
	}
	return items
}

func TestTokenMessengerGet(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNTokenMessengers(cctpKeeper, ctx, 10)
	for _, item := range items {
		rst, found := cctpKeeper.GetTokenMessenger(ctx,
			item.tokenMessenger.DomainId,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item.tokenMessenger),
			nullify.Fill(&rst),
		)
	}
}

func TestTokenMessengerRemove(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNTokenMessengers(cctpKeeper, ctx, 10)
	for _, item := range items {
		cctpKeeper.DeleteTokenMessenger(ctx, item.address)
		_, found := cctpKeeper.GetTokenMessenger(ctx, item.address)
		require.False(t, found)
	}
}

func TestTokenMessengersGetAll(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNTokenMessengers(cctpKeeper, ctx, 10)
	denom := make([]types.TokenMessenger, len(items))
	for i, item := range items {
		denom[i] = item.tokenMessenger
	}
	require.ElementsMatch(t,
		nullify.Fill(denom),
		nullify.Fill(cctpKeeper.GetAllTokenMessengers(ctx)),
	)
}
