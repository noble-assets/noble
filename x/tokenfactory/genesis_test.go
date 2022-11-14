package tokenfactory_test

import (
	"testing"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/tokenfactory"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	"github.com/stretchr/testify/require"
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
				Minter: "0",
			},
			{
				Minter: "1",
			},
		},
		MintingDenom: &types.MintingDenom{
			Denom: "65",
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
	require.Equal(t, genesisState.MintingDenom, got.MintingDenom)
	// this line is used by starport scaffolding # genesis/test/assert
}
