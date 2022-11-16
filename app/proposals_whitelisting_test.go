package app_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	icssimapp "github.com/cosmos/interchain-security/testutil/simapp"
	"github.com/strangelove-ventures/noble/app"
	"github.com/strangelove-ventures/noble/cmd"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmdb "github.com/tendermint/tm-db"
)

func TestConsumerWhitelistingKeys(t *testing.T) {
	chain := ibctesting.NewTestChain(t, icssimapp.NewBasicCoordinator(t), SetupTestingAppConsumer, "test")
	paramKeeper := chain.App.(*app.App).ParamsKeeper
	for paramKey := range app.WhitelistedParams {
		ss, ok := paramKeeper.GetSubspace(paramKey.Subspace)
		require.True(t, ok, "Unknown subspace %s", paramKey.Subspace)
		hasKey := ss.Has(chain.GetContext(), []byte(paramKey.Key))
		require.True(t, hasKey, "Invalid key %s for subspace %s", paramKey.Key, paramKey.Subspace)
	}
}

func SetupTestingAppConsumer() (ibctesting.TestingApp, map[string]json.RawMessage) {
	encoding := cmd.MakeEncodingConfig(app.ModuleBasics)
	testApp := app.New(
		log.NewNopLogger(),
		tmdb.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		simapp.EmptyAppOptions{},
	)

	return testApp.(*app.App), app.NewDefaultGenesisState(encoding.Marshaler)
}
