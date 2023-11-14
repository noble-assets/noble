package keeper_test

import (
	"testing"

	testkeeper "github.com/strangelove-ventures/noble/v4/testutil/keeper"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.StableTokenFactoryKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
