package cmd

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CosmosApp implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type CosmosApp interface {
	// Name is the assigned name of the app.
	Name() string

	// The application types codec.
	// NOTE: This should be sealed before being returned.
	LegacyAmino() *codec.LegacyAmino

	// BeginBlocker updates every begin block.
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

	// EndBlocker updates every end block.
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock

	// InitChainer updates at chain (i.e app) initialization.
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain

	// LoadHeight loads the app at a given height.
	LoadHeight(height int64) error

	// ExportAppStateAndValidators exports the state of the application for a genesis file.
	ExportAppStateAndValidators(
		forZeroHeight bool, jailAllowedAddrs []string,
	) (types.ExportedApp, error)

	// ModuleAccountAddrs are registered module account addreses.
	ModuleAccountAddrs() map[string]bool
}
