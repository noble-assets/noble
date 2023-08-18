package globalfee

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	// Auth
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// GlobalFee
	moduleV1 "github.com/strangelove-ventures/noble/pulsar/noble/globalfee/module/v1"
	"github.com/strangelove-ventures/noble/x/globalfee/client/cli"
	"github.com/strangelove-ventures/noble/x/globalfee/keeper"
	"github.com/strangelove-ventures/noble/x/globalfee/types"
	// Gov
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ appmodule.AppModule   = AppModule{}
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ------------------------------ AppModuleBasic -------------------------------

type AppModuleBasic struct{}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

func (a AppModuleBasic) GetQueryCmd() *cobra.Command { return cli.GetQueryCmd() }

func (a AppModuleBasic) Name() string { return types.ModuleName }

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	_ = types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

func (a AppModuleBasic) RegisterInterfaces(registry codecTypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var gs types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal x/%s genesis state: %w", types.ModuleName, err)
	}

	return gs.Validate()
}

// --------------------------------- AppModule ---------------------------------

type AppModule struct {
	AppModuleBasic
	keeper *keeper.Keeper
}

func (AppModule) BeginBlock(_ sdk.Context, _ *abci.RequestBeginBlock) {}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (AppModule) EndBlock(_ sdk.Context, _ *abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (a AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, a.keeper)
	return cdc.MustMarshalJSON(gs)
}

func (a AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var gs types.GenesisState
	cdc.MustUnmarshalJSON(data, &gs)

	InitGenesis(ctx, a.keeper, gs)

	return []abci.ValidatorUpdate{}
}

func (a AppModule) IsOnePerModuleType() {}

func (a AppModule) IsAppModule() {}

func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), a.keeper)
	types.RegisterQueryServer(cfg.QueryServer(), a.keeper)
}

// ------------------------------ App Wiring Setup -----------------------------

func init() {
	appmodule.Register(&moduleV1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type Inputs struct {
	depinject.In

	Config           *moduleV1.Module
	Cdc              codec.Codec
	Key              *storeTypes.KVStoreKey
	MsgServiceRouter *baseapp.MsgServiceRouter
}

type Outputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in Inputs) Outputs {
	authority := authTypes.NewModuleAddress(govTypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authTypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	globalFeeKeeper := keeper.NewKeeper(
		in.Cdc,
		in.Key,
		authority.String(),
	)
	globalFeeModule := AppModule{AppModuleBasic{}, globalFeeKeeper}

	return Outputs{Keeper: globalFeeKeeper, Module: globalFeeModule}
}
