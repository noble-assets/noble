package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/strangelove-ventures/noble/x/poa/types"
)

func TestKeeperParamsFunctions(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	// SetParams test
	keeper.SetParams(ctx, types.DefaultParams())

	// GetParams test
	params := keeper.GetParams(ctx)

	require.Equal(t, uint32(49), params.Quorum)
	require.Equal(t, uint32(100), params.MaxValidators)
}
