package app

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/circlefin/noble-fiattokenfactory/x/blockibc"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	soloclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	authoritytypes "github.com/noble-assets/authority/x/authority/types"
	"github.com/noble-assets/forwarding/v2/x/forwarding"
)

func (app *NobleApp) RegisterIBCModules() error {
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(capabilitytypes.StoreKey),
		storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey),
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(icahosttypes.StoreKey),
		storetypes.NewKVStoreKey(transfertypes.StoreKey),
		storetypes.NewKVStoreKey(packetforwardtypes.StoreKey),
	); err != nil {
		return err
	}

	app.ParamsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(clienttypes.ParamKeyTable().RegisterParamSet(&connectiontypes.Params{}))
	app.ParamsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(transfertypes.ModuleName).WithKeyTable(transfertypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		app.appCodec,
		app.GetKey(capabilitytypes.StoreKey),
		app.GetMemKey(capabilitytypes.MemStoreKey),
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.IBCKeeper = ibckeeper.NewKeeper(
		app.appCodec,
		app.GetKey(ibcexported.StoreKey),
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		authoritytypes.ModuleAddress.String(),
	)
	app.ScopedIBCKeeper = scopedIBCKeeper

	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(icahosttypes.StoreKey),
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
		authoritytypes.ModuleAddress.String(),
	)
	app.ScopedICAHostKeeper = scopedICAHostKeeper

	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(transfertypes.ModuleName)
	app.TransferKeeper = transferkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(transfertypes.StoreKey),
		app.GetSubspace(transfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		authoritytypes.ModuleAddress.String(),
	)
	app.ScopedTransferKeeper = scopedTransferKeeper

	app.PFMKeeper = packetforwardkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(packetforwardtypes.StoreKey),
		app.TransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.DistributionKeeper,
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		authoritytypes.ModuleAddress.String(),
	)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = forwarding.NewMiddleware(transferStack, app.AccountKeeper, app.ForwardingKeeper)
	transferStack = packetforward.NewIBCMiddleware(transferStack, app.PFMKeeper, 0, packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp)
	transferStack = blockibc.NewIBCMiddleware(transferStack, app.FiatTokenFactoryKeeper)

	ibcRouter := porttypes.NewRouter().
		AddRoute(transfertypes.ModuleName, transferStack).
		AddRoute(icahosttypes.SubModuleName, icahost.NewIBCModule(app.ICAHostKeeper))
	app.IBCKeeper.SetRouter(ibcRouter)

	app.ForwardingKeeper.SetIBCKeepers(app.IBCKeeper.ChannelKeeper, app.TransferKeeper)

	if err := app.RegisterModules(
		capability.NewAppModule(app.appCodec, *app.CapabilityKeeper, false),
		ibc.NewAppModule(app.IBCKeeper),
		ica.NewAppModule(nil, &app.ICAHostKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		packetforward.NewAppModule(app.PFMKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		tmclient.NewAppModule(),
		soloclient.NewAppModule(),
	); err != nil {
		return err
	}

	return nil
}
