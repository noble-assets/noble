package keeper_test

import (
	"testing"

	testkeeper "noble/testutil/keeper"
	"noble/x/tokenfactory/types"

	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	ctx, tk, _ := testkeeper.NewTestSetup(t)
	params := types.DefaultParams()

	tk.TokenfactoryKeeper.SetParams(ctx, params)

	require.EqualValues(t, params, tk.TokenfactoryKeeper.GetParams(ctx))
}
