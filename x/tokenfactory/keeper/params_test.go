package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "noble/testutil/keeper"
	"noble/x/tokenfactory/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.TokenfactoryKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
