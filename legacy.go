// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package noble

import (
	storetypes "cosmossdk.io/store/types"
	"dollar.noble.xyz/v2"
	"github.com/circlefin/noble-fiattokenfactory/x/blockibc"
	"github.com/cosmos/cosmos-sdk/runtime"
	pfm "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	pfmkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v8"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
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
	authoritytypes "github.com/noble-assets/authority/types"
	"github.com/noble-assets/forwarding/v2"
	"github.com/noble-assets/wormhole"
	wormholetypes "github.com/noble-assets/wormhole/types"
)

func (app *App) RegisterLegacyModules() error {
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(capabilitytypes.StoreKey),
		storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey),
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(icahosttypes.StoreKey),
		storetypes.NewKVStoreKey(pfmtypes.StoreKey),
		storetypes.NewKVStoreKey(ratelimittypes.StoreKey),
		storetypes.NewKVStoreKey(transfertypes.StoreKey),
	); err != nil {
		return err
	}

	app.ParamsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(clienttypes.ParamKeyTable().RegisterParamSet(&connectiontypes.Params{}))
	app.ParamsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(ratelimittypes.ModuleName).WithKeyTable(ratelimittypes.ParamKeyTable())
	app.ParamsKeeper.Subspace(transfertypes.ModuleName).WithKeyTable(transfertypes.ParamKeyTable())

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
	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())

	// Create custom ICS4Wrapper so that we can block outgoing $USDN IBC transfers.
	ics4Wrapper := dollar.NewICS4Wrapper(app.IBCKeeper.ChannelKeeper, app.DollarKeeper)

	app.RateLimitKeeper = *ratelimitkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(ratelimittypes.StoreKey)),
		app.GetSubspace(ratelimittypes.ModuleName),
		authoritytypes.ModuleAddress.String(),
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		ics4Wrapper,
	)

	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(transfertypes.ModuleName)
	app.TransferKeeper = transferkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(transfertypes.StoreKey),
		app.GetSubspace(transfertypes.ModuleName),
		app.RateLimitKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		authoritytypes.ModuleAddress.String(),
	)
	app.PFMKeeper = pfmkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(pfmtypes.StoreKey),
		app.TransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.BankKeeper,
		app.IBCKeeper.ChannelKeeper,
		authoritytypes.ModuleAddress.String(),
	)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = ratelimit.NewIBCMiddleware(app.RateLimitKeeper, transferStack)
	transferStack = forwarding.NewMiddleware(transferStack, app.AccountKeeper, app.ForwardingKeeper)
	transferStack = pfm.NewIBCMiddleware(
		transferStack,
		app.PFMKeeper,
		0,
		pfmkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)
	transferStack = blockibc.NewIBCMiddleware(transferStack, app.FTFKeeper)

	ibcRouter := porttypes.NewRouter().
		AddRoute(icahosttypes.SubModuleName, icahost.NewIBCModule(app.ICAHostKeeper)).
		AddRoute(transfertypes.ModuleName, transferStack).
		AddRoute(wormholetypes.ModuleName, wormhole.NewIBCModule(app.WormholeKeeper))
	app.IBCKeeper.SetRouter(ibcRouter)

	app.DollarKeeper.SetIBCKeepers(app.IBCKeeper.ChannelKeeper, app.TransferKeeper)

	app.ForwardingKeeper.SetIBCKeepers(app.IBCKeeper.ChannelKeeper, app.TransferKeeper)

	scopedWormholeKeeper := app.CapabilityKeeper.ScopeToModule(wormholetypes.ModuleName)
	app.WormholeKeeper.SetIBCKeepers(app.IBCKeeper.ChannelKeeper, app.IBCKeeper.PortKeeper, scopedWormholeKeeper)

	return app.RegisterModules(
		capability.NewAppModule(app.appCodec, *app.CapabilityKeeper, true),
		ibc.NewAppModule(app.IBCKeeper),
		ica.NewAppModule(nil, &app.ICAHostKeeper),
		pfm.NewAppModule(app.PFMKeeper, app.GetSubspace(pfmtypes.ModuleName)),
		transfer.NewAppModule(app.TransferKeeper),
		tmclient.NewAppModule(),
		soloclient.NewAppModule(),
		ratelimit.NewAppModule(app.appCodec, app.RateLimitKeeper),
	)
}
