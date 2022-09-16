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
		Pauser: &types.Pauser{
			Address: "96",
		},
		Blacklister: &types.Blacklister{
			Address: "20",
		},
		Owner: &types.Owner{
			Address: "98",
		},
		Admin: &types.Admin{
			Address: "45",
		},
		MinterControllerList: []types.MinterController{
			{
				MinterAddress: "0",
			},
			{
				MinterAddress: "1",
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
	require.Equal(t, genesisState.Pauser, got.Pauser)
	require.Equal(t, genesisState.Blacklister, got.Blacklister)
	require.Equal(t, genesisState.Owner, got.Owner)
	require.Equal(t, genesisState.Admin, got.Admin)
	require.ElementsMatch(t, genesisState.MinterControllerList, got.MinterControllerList)
	// this line is used by starport scaffolding # genesis/test/assert
}
