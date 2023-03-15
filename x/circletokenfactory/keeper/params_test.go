package keeper_test

import (
	"testing"

	testkeeper "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/x/circletokenfactory/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.CircleTokenfactoryKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
