package app_test

import (
	"os"
	"testing"

	"github.com/strangelove-ventures/noble/v5/app"
	"github.com/strangelove-ventures/noble/v5/cmd"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func init() {
	simapp.GetSimulatorFlags()
}

type SimApp interface {
	cmd.App
	GetBaseApp() *baseapp.BaseApp
	AppCodec() codec.Codec
	SimulationManager() *module.SimulationManager
	ModuleAccountAddrs() map[string]bool
	Name() string
	LegacyAmino() *codec.LegacyAmino
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `starport chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50 -Commit=true -Verbose=true -Enabled=true`
func BenchmarkSimulation(b *testing.B) {
	simapp.FlagNumBlocksValue = 250
	simapp.FlagBlockSizeValue = 50
	simapp.FlagCommitValue = true
	simapp.FlagVerboseValue = true
	simapp.FlagEnabledValue = true

	config, db, dir, logger, _, err := simapp.SetupSimulation("goleveldb-app-sim", "Simulation")
	require.NoError(b, err, "simulation setup failed")

	b.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		require.NoError(b, err)
	})

	encoding := cmd.MakeEncodingConfig(app.ModuleBasics)

	app := app.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		simapp.EmptyAppOptions{},
	)

	simApp, ok := app.(SimApp)
	require.True(b, ok, "can't use simapp")

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		simApp.GetBaseApp(),
		simapp.AppStateFn(simApp.AppCodec(), simApp.SimulationManager()),
		simulationtypes.RandomAccounts,
		simapp.SimulationOperations(simApp, simApp.AppCodec(), config),
		simApp.ModuleAccountAddrs(),
		config,
		simApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(simApp, config, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}
}
