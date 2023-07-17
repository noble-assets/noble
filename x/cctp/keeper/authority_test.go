package keeper_test

import (
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
)

func TestAuthorityGet(t *testing.T) {
	keeper, ctx := keepertest.CctpKeeper(t)

	Authority := types.Authority{Address: "1"}
	keeper.SetAuthority(ctx, Authority)

	rst, found := keeper.GetAuthority(ctx)
	require.True(t, found)
	require.Equal(t,
		Authority,
		nullify.Fill(&rst),
	)

	newAuthority := types.Authority{Address: "2"}

	keeper.SetAuthority(ctx, newAuthority)

	rst, found = keeper.GetAuthority(ctx)
	require.True(t, found)
	require.Equal(t,
		newAuthority,
		nullify.Fill(&rst),
	)
}
