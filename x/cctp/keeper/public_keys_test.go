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

type publicKeyWrapper struct {
	address   string
	publicKey types.PublicKeys
}

func createNPublicKeys(keeper *keeper.Keeper, ctx sdk.Context, n int) []publicKeyWrapper {
	items := make([]publicKeyWrapper, n)
	for i := range items {
		items[i].address = sample.AccAddress()
		items[i].publicKey.Key = "PublicKey" + strconv.Itoa(i)

		keeper.SetPublicKey(ctx, items[i].publicKey)
	}
	return items
}

func TestPublicKeysGet(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNPublicKeys(cctpKeeper, ctx, 10)
	for _, item := range items {
		rst, found := cctpKeeper.GetPublicKey(ctx,
			item.publicKey.Key,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item.publicKey),
			nullify.Fill(&rst),
		)
	}
}

func TestPublicKeysRemove(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNPublicKeys(cctpKeeper, ctx, 10)
	for _, item := range items {
		cctpKeeper.DeletePublicKey(ctx, item.address)
		_, found := cctpKeeper.GetPublicKey(ctx, item.address)
		require.False(t, found)
	}
}

func TestPublicKeysGetAll(t *testing.T) {
	cctpKeeper, ctx := keepertest.CctpKeeper(t)
	items := createNPublicKeys(cctpKeeper, ctx, 10)
	denom := make([]types.PublicKeys, len(items))
	for i, item := range items {
		denom[i] = item.publicKey
	}
	require.ElementsMatch(t,
		nullify.Fill(denom),
		nullify.Fill(cctpKeeper.GetAllPublicKeys(ctx)),
	)
}
