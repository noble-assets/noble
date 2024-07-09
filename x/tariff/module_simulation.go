package tariff

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/noble-assets/noble/v6/x/tariff/types"
	"math/rand"
)

var _ module.AppModuleSimulation = AppModule{}

func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}

func (am AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	return nil
}

func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
