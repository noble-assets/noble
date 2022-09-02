package tokenfactory_test

import (
	"testing"

	testkeeper "noble/testutil/keeper"
	"noble/testutil/nullify"
	"noble/x/tokenfactory"
	"noble/x/tokenfactory/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	ctx, tk, _ := testkeeper.NewTestSetup(t)
	tokenfactory.InitGenesis(ctx, *tk.TokenfactoryKeeper, genesisState)
	got := tokenfactory.ExportGenesis(ctx, *tk.TokenfactoryKeeper)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
