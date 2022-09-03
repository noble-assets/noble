package tokenfactory_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "noble/testutil/keeper"
	"noble/testutil/nullify"
	"noble/x/tokenfactory"
	"noble/x/tokenfactory/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		BlacklistedList: []types.Blacklisted{
			{
				Address: "0",
			},
			{
				Address: "1",
			},
		},
		Paused: &types.Paused{
			Paused: true,
		},
		MasterMinter: &types.MasterMinter{
			Address: "79",
		},
		MintersList: []types.Minters{
			{
				Address: "0",
			},
			{
				Address: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.TokenfactoryKeeper(t)
	tokenfactory.InitGenesis(ctx, *k, genesisState)
	got := tokenfactory.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.BlacklistedList, got.BlacklistedList)
	require.Equal(t, genesisState.Paused, got.Paused)
	require.Equal(t, genesisState.MasterMinter, got.MasterMinter)
	require.ElementsMatch(t, genesisState.MintersList, got.MintersList)
	// this line is used by starport scaffolding # genesis/test/assert
}
