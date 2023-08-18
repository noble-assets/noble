package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	// BlockIBC
	blockIBC "github.com/strangelove-ventures/noble/x/blockibc"
	// IBC Client
	ibcClientSolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibcClientTendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	// IBC Core
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcPortTypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcTypes "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibcKeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	// IBC Fee
	ibcFee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcFeeKeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	ibcFeeTypes "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/types"
	// IBC Transfer
	ibcTransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibcTransferKeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibcTransferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	// ICA
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	// ICA Controller
	icaController "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller"
	icaControllerKeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	icaControllerTypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	// ICA Host
	icaHost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icaHostKeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icaHostTypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	// PFM
	pfm "github.com/strangelove-ventures/packet-forward-middleware/v7/router"
	pfmKeeper "github.com/strangelove-ventures/packet-forward-middleware/v7/router/keeper"
	pfmTypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"
)

var (
	keys = sdk.NewKVStoreKeys(
		ibcTypes.StoreKey, ibcFeeTypes.StoreKey, ibcTransferTypes.StoreKey,
		icaControllerTypes.StoreKey, icaHostTypes.StoreKey, pfmTypes.StoreKey,
	)

	memKeys = sdk.NewMemoryStoreKeys()

	subspaces = []string{
		ibcTypes.ModuleName, ibcFeeTypes.ModuleName, ibcTransferTypes.ModuleName,
		icaControllerTypes.SubModuleName, icaHostTypes.SubModuleName, pfmTypes.ModuleName,
	}
)

func (app *NobleApp) RegisterLegacyModules() {
	// Register the additional keys.
	app.MountKVStores(keys)
	app.MountMemoryStores(memKeys)

	// Initialise the additional param subspaces.
	for _, subspace := range subspaces {
		app.ParamsKeeper.Subspace(subspace)
	}

	// Keeper: IBC
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcTypes.ModuleName)
	app.IBCKeeper = ibcKeeper.NewKeeper(
		app.appCodec,
		keys[ibcTypes.StoreKey],
		app.GetSubspace(ibcTypes.ModuleName),

		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)
	app.ScopedIBCKeeper = scopedIBCKeeper

	// Keeper: IBC Fee
	app.IBCFeeKeeper = ibcFeeKeeper.NewKeeper(
		app.appCodec,
		keys[ibcFeeTypes.ModuleName],

		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,

		app.AccountKeeper,
		app.BankKeeper,
	)

	// Keeper: PFM
	app.TariffKeeper.SetICS4Wrapper(app.IBCKeeper.ChannelKeeper)
	app.PFMKeeper = pfmKeeper.NewKeeper(
		app.appCodec,
		keys[pfmTypes.StoreKey],
		app.GetSubspace(pfmTypes.ModuleName),
		app.IBCTransferKeeper, // will be zero-value here. reference set later on with SetTransferKeeper.
		app.IBCKeeper.ChannelKeeper,
		app.DistributionKeeper,
		app.BankKeeper,
		app.TariffKeeper,
	)

	// Keeper: IBC Transfer
	scopedIBCTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibcTransferTypes.ModuleName)
	app.IBCTransferKeeper = ibcTransferKeeper.NewKeeper(
		app.appCodec,
		keys[ibcTransferTypes.StoreKey],
		app.GetSubspace(ibcTransferTypes.ModuleName),

		app.PFMKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,

		app.AccountKeeper,
		app.BankKeeper,
		scopedIBCTransferKeeper,
	)
	app.ScopedIBCTransferKeeper = scopedIBCTransferKeeper
	app.PFMKeeper.SetTransferKeeper(app.IBCTransferKeeper)

	// Keeper: ICA Controller
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icaControllerTypes.SubModuleName)
	app.ICAControllerKeeper = icaControllerKeeper.NewKeeper(
		app.appCodec,
		keys[icaControllerTypes.StoreKey],
		app.GetSubspace(icaControllerTypes.SubModuleName),

		app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,

		scopedICAControllerKeeper,
		app.MsgServiceRouter(),
	)
	app.ScopedICAControllerKeeper = scopedICAControllerKeeper

	// Keeper: ICA Host
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icaHostTypes.SubModuleName)
	app.ICAHostKeeper = icaHostKeeper.NewKeeper(
		app.appCodec,
		keys[icaHostTypes.StoreKey],
		app.GetSubspace(icaHostTypes.SubModuleName),

		app.IBCFeeKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,

		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
	)
	app.ScopedICAHostKeeper = scopedICAHostKeeper

	// IBC: Create a router.
	var ibcTransferStack ibcPortTypes.IBCModule
	ibcTransferStack = ibcTransfer.NewIBCModule(app.IBCTransferKeeper)
	ibcTransferStack = pfm.NewIBCMiddleware(
		ibcTransferStack,
		app.PFMKeeper,
		0,
		pfmKeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		pfmKeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)
	ibcTransferStack = blockIBC.NewIBCMiddleware(ibcTransferStack, app.TokenFactoryKeeper, app.FiatTokenFactoryKeeper)
	ibcTransferStack = ibcFee.NewIBCMiddleware(ibcTransferStack, app.IBCFeeKeeper)

	var icaControllerStack ibcPortTypes.IBCModule
	icaControllerStack = icaController.NewIBCMiddleware(icaControllerStack, app.ICAControllerKeeper)
	icaControllerStack = ibcFee.NewIBCMiddleware(icaControllerStack, app.IBCFeeKeeper)

	var icaHostStack ibcPortTypes.IBCModule
	icaHostStack = icaHost.NewIBCModule(app.ICAHostKeeper)
	icaHostStack = ibcFee.NewIBCMiddleware(icaHostStack, app.IBCFeeKeeper)

	ibcRouter := ibcPortTypes.NewRouter()
	ibcRouter.AddRoute(ibcTransferTypes.ModuleName, ibcTransferStack).
		AddRoute(icaControllerTypes.SubModuleName, icaControllerStack).
		AddRoute(icaHostTypes.SubModuleName, icaHostStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	// Register modules and interfaces/services.
	legacyModules := []module.AppModule{
		ibc.NewAppModule(app.IBCKeeper),
		ibcFee.NewAppModule(app.IBCFeeKeeper),
		ibcTransfer.NewAppModule(app.IBCTransferKeeper),
		ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper),
		pfm.NewAppModule(app.PFMKeeper),
	}
	if err := app.RegisterModules(legacyModules...); err != nil {
		panic(err)
	}

	for _, m := range legacyModules {
		if s, ok := m.(module.HasServices); ok {
			s.RegisterServices(app.Configurator())
		}
	}

	ibcClientSolomachine.AppModuleBasic{}.RegisterInterfaces(app.interfaceRegistry)
	ibcClientTendermint.AppModuleBasic{}.RegisterInterfaces(app.interfaceRegistry)
}
