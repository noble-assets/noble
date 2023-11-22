package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

<<<<<<< HEAD
	keepertest "github.com/strangelove-ventures/noble/v4/testutil/keeper"
	"github.com/strangelove-ventures/noble/v4/testutil/nullify"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	keepertest "github.com/noble-assets/noble/v5/testutil/keeper"
	"github.com/noble-assets/noble/v5/testutil/nullify"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))
)

func TestOwnerGet(t *testing.T) {
	keeper, ctx := keepertest.StableTokenFactoryKeeper(t)

	owner := types.Owner{Address: "1"}
	keeper.SetOwner(ctx, owner)

	rst, found := keeper.GetOwner(ctx)
	require.True(t, found)
	require.Equal(t,
		owner,
		nullify.Fill(&rst),
	)

	newOwner := types.Owner{Address: "2"}

	keeper.SetPendingOwner(ctx, newOwner)

	rst, found = keeper.GetPendingOwner(ctx)
	require.True(t, found)
	require.Equal(t,
		newOwner,
		nullify.Fill(&rst),
	)
}
