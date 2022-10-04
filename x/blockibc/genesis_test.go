package blockibc_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "noble/testutil/keeper"
	"noble/testutil/nullify"
	"noble/x/blockibc"
	"noble/x/blockibc/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.BlockibcKeeper(t)
	blockibc.InitGenesis(ctx, *k, genesisState)
	got := blockibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	// this line is used by starport scaffolding # genesis/test/assert
}
