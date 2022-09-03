package tokenfactory

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"noble/testutil/sample"
	tokenfactorysimulation "noble/x/tokenfactory/simulation"
	"noble/x/tokenfactory/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = tokenfactorysimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgChangeAdmin = "op_weight_msg_change_admin"
	// TODO: Determine the simulation weight value
	defaultWeightMsgChangeAdmin int = 100

	opWeightMsgUpdateMasterMinter = "op_weight_msg_update_master_minter"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateMasterMinter int = 100

	opWeightMsgUpdatePauser = "op_weight_msg_update_pauser"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdatePauser int = 100

	opWeightMsgUpdateBlacklister = "op_weight_msg_update_blacklister"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateBlacklister int = 100

	opWeightMsgUpdateOwner = "op_weight_msg_update_owner"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateOwner int = 100

	opWeightMsgConfigureMinter = "op_weight_msg_configure_minter"
	// TODO: Determine the simulation weight value
	defaultWeightMsgConfigureMinter int = 100

	opWeightMsgRemoveMinter = "op_weight_msg_remove_minter"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRemoveMinter int = 100

	opWeightMsgMint = "op_weight_msg_mint"
	// TODO: Determine the simulation weight value
	defaultWeightMsgMint int = 100

	opWeightMsgBurn = "op_weight_msg_burn"
	// TODO: Determine the simulation weight value
	defaultWeightMsgBurn int = 100

	opWeightMsgBlacklist = "op_weight_msg_blacklist"
	// TODO: Determine the simulation weight value
	defaultWeightMsgBlacklist int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	tokenfactoryGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&tokenfactoryGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgChangeAdmin int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgChangeAdmin, &weightMsgChangeAdmin, nil,
		func(_ *rand.Rand) {
			weightMsgChangeAdmin = defaultWeightMsgChangeAdmin
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgChangeAdmin,
		tokenfactorysimulation.SimulateMsgChangeAdmin(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateMasterMinter int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateMasterMinter, &weightMsgUpdateMasterMinter, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateMasterMinter = defaultWeightMsgUpdateMasterMinter
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateMasterMinter,
		tokenfactorysimulation.SimulateMsgUpdateMasterMinter(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdatePauser int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdatePauser, &weightMsgUpdatePauser, nil,
		func(_ *rand.Rand) {
			weightMsgUpdatePauser = defaultWeightMsgUpdatePauser
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdatePauser,
		tokenfactorysimulation.SimulateMsgUpdatePauser(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateBlacklister int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateBlacklister, &weightMsgUpdateBlacklister, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateBlacklister = defaultWeightMsgUpdateBlacklister
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateBlacklister,
		tokenfactorysimulation.SimulateMsgUpdateBlacklister(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateOwner int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateOwner, &weightMsgUpdateOwner, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateOwner = defaultWeightMsgUpdateOwner
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateOwner,
		tokenfactorysimulation.SimulateMsgUpdateOwner(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgConfigureMinter int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgConfigureMinter, &weightMsgConfigureMinter, nil,
		func(_ *rand.Rand) {
			weightMsgConfigureMinter = defaultWeightMsgConfigureMinter
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgConfigureMinter,
		tokenfactorysimulation.SimulateMsgConfigureMinter(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRemoveMinter int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgRemoveMinter, &weightMsgRemoveMinter, nil,
		func(_ *rand.Rand) {
			weightMsgRemoveMinter = defaultWeightMsgRemoveMinter
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRemoveMinter,
		tokenfactorysimulation.SimulateMsgRemoveMinter(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgMint int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgMint, &weightMsgMint, nil,
		func(_ *rand.Rand) {
			weightMsgMint = defaultWeightMsgMint
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMint,
		tokenfactorysimulation.SimulateMsgMint(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgBurn int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgBurn, &weightMsgBurn, nil,
		func(_ *rand.Rand) {
			weightMsgBurn = defaultWeightMsgBurn
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgBurn,
		tokenfactorysimulation.SimulateMsgBurn(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgBlacklist int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgBlacklist, &weightMsgBlacklist, nil,
		func(_ *rand.Rand) {
			weightMsgBlacklist = defaultWeightMsgBlacklist
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgBlacklist,
		tokenfactorysimulation.SimulateMsgBlacklist(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
