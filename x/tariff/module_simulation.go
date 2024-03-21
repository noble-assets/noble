package tariff

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/noble-assets/noble/v5/x/tariff/types"
)

func (am AppModule) GenerateGenesisState(simState *module.SimulationState) {
	genesis := types.DefaultGenesis()
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}

func (am AppModule) RegisterStoreDecoder(_ simulation.StoreDecoderRegistry) {}

func (am AppModule) WeightedOperations(_ module.SimulationState) []simtypes.WeightedOperation {
	return nil
}
