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
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/libs/bytes"
	cmtos "github.com/cometbft/cometbft/libs/os"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/runtime"
	serverapi "github.com/cosmos/cosmos-sdk/server/api"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/noble-assets/noble/v11/api"
	"github.com/noble-assets/noble/v11/jester"
	"github.com/noble-assets/noble/v11/upgrade"
	"github.com/spf13/cast"

	_ "cosmossdk.io/x/evidence"
	_ "cosmossdk.io/x/feegrant/module"
	_ "cosmossdk.io/x/upgrade"
	_ "dollar.noble.xyz/v2"
	_ "github.com/bcp-innovations/hyperlane-cosmos/x/core"
	_ "github.com/bcp-innovations/hyperlane-cosmos/x/warp"
	_ "github.com/circlefin/noble-cctp/x/cctp"
	_ "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	_ "github.com/cosmos/cosmos-sdk/x/authz/module"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/crisis"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/slashing"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
	_ "github.com/monerium/module-noble/v2"
	_ "github.com/noble-assets/authority"
	_ "github.com/noble-assets/forwarding/v2"
	"github.com/noble-assets/globalfee"
	_ "github.com/noble-assets/halo/v2"
	_ "github.com/noble-assets/orbiter"
	_ "github.com/noble-assets/wormhole"
	_ "github.com/ondoprotocol/usdy-noble/v2"
	_ "swap.noble.xyz"

	// Cosmos Modules
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	// IBC Modules
	pfmkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	ratelimitkeeper "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	// Circle Modules
	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	ftfkeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"

	// Ondo Modules
	aurakeeper "github.com/ondoprotocol/usdy-noble/v2/keeper"

	// Hashnote Modules
	halokeeper "github.com/noble-assets/halo/v2/keeper"

	// Hyperlane Modules
	hyperlanekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"

	// Monerium Modules
	florinkeeper "github.com/monerium/module-noble/v2/keeper"

	// Noble Modules
	dollarkeeper "dollar.noble.xyz/v2/keeper"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	forwardingkeeper "github.com/noble-assets/forwarding/v2/keeper"
	globalfeekeeper "github.com/noble-assets/globalfee/keeper"
	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	wormholekeeper "github.com/noble-assets/wormhole/keeper"
	swapkeeper "swap.noble.xyz/keeper"
)

var DefaultNodeHome string

//go:embed app.yaml
var AppConfigYAML []byte

var (
	_ runtime.AppI            = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// App defines the interface of Noble's Cosmos SDK-based application that extends the default ABCI interface.
type App struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// Cosmos Modules
	AccountKeeper   authkeeper.AccountKeeper
	AuthzKeeper     authzkeeper.Keeper
	BankKeeper      bankkeeper.Keeper
	ConsensusKeeper consensuskeeper.Keeper
	CrisisKeeper    *crisiskeeper.Keeper
	EvidenceKeeper  evidencekeeper.Keeper
	FeeGrantKeeper  feegrantkeeper.Keeper
	ParamsKeeper    paramskeeper.Keeper
	SlashingKeeper  slashingkeeper.Keeper
	StakingKeeper   *stakingkeeper.Keeper
	UpgradeKeeper   *upgradekeeper.Keeper
	// IBC Modules
	CapabilityKeeper *capabilitykeeper.Keeper
	IBCKeeper        *ibckeeper.Keeper
	ICAHostKeeper    icahostkeeper.Keeper
	PFMKeeper        *pfmkeeper.Keeper
	RateLimitKeeper  ratelimitkeeper.Keeper
	TransferKeeper   transferkeeper.Keeper
	// Circle Modules
	CCTPKeeper *cctpkeeper.Keeper
	FTFKeeper  *ftfkeeper.Keeper
	// Ondo Modules
	AuraKeeper *aurakeeper.Keeper
	// Hashnote Modules
	HaloKeeper *halokeeper.Keeper
	// Hyperlane Modules
	HyperlaneKeeper *hyperlanekeeper.Keeper
	WarpKeeper      warpkeeper.Keeper
	// Monerium Modules
	FlorinKeeper *florinkeeper.Keeper
	// Noble Modules
	AuthorityKeeper  *authoritykeeper.Keeper
	DollarKeeper     *dollarkeeper.Keeper
	ForwardingKeeper *forwardingkeeper.Keeper
	GlobalFeeKeeper  *globalfeekeeper.Keeper
	OrbiterKeeper    *orbiterkeeper.Keeper
	SwapKeeper       *swapkeeper.Keeper
	WormholeKeeper   *wormholekeeper.Keeper
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".noble")
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appconfig.LoadYAML(AppConfigYAML),
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
			},
		),
	)
}

// NewApp initializes and returns a new Noble application.
func NewApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*App, error) {
	var (
		app        = &App{}
		appBuilder *runtime.AppBuilder
	)

	if err := depinject.Inject(
		depinject.Configs(
			AppConfig(),
			depinject.Supply(
				logger,
				appOpts,
			),
		),
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		// Cosmos Modules
		&app.AccountKeeper,
		&app.AuthzKeeper,
		&app.BankKeeper,
		&app.ConsensusKeeper,
		&app.CrisisKeeper,
		&app.EvidenceKeeper,
		&app.FeeGrantKeeper,
		&app.ParamsKeeper,
		&app.SlashingKeeper,
		&app.StakingKeeper,
		&app.UpgradeKeeper,
		// Circle Modules
		&app.CCTPKeeper,
		&app.FTFKeeper,
		// Hashnote Modules
		&app.HaloKeeper,
		// Hyperlane Modules
		&app.HyperlaneKeeper,
		&app.WarpKeeper,
		// Monerium Modules
		&app.FlorinKeeper,
		// Ondo Modules
		&app.AuraKeeper,
		// Noble Modules
		&app.AuthorityKeeper,
		&app.DollarKeeper,
		&app.ForwardingKeeper,
		&app.GlobalFeeKeeper,
		&app.OrbiterKeeper,
		&app.SwapKeeper,
		&app.WormholeKeeper,
	); err != nil {
		return nil, err
	}

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	app.RegisterOrbiterControllers()

	if err := app.RegisterLegacyModules(); err != nil {
		return nil, err
	}

	// When initializing the upgrade keeper via dependency injection, the
	// initial module version map is created using only the modules that are
	// wired through dependency injection. As a result, any "legacy" modules
	// (those that don't support dependency injection) are excluded. The line
	// below updates the version map to ensure that all modules are included.
	app.UpgradeKeeper.SetInitVersionMap(app.ModuleManager.GetVersionMap())

	anteHandler, err := NewAnteHandler(HandlerOptions{
		HandlerOptions: ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			FeegrantKeeper:  app.FeeGrantKeeper,
			SignModeHandler: app.txConfig.SignModeHandler(),
			TxFeeChecker:    globalfee.TxFeeChecker(app.GlobalFeeKeeper),
			SigGasConsumer:  SigVerificationGasConsumer,
		},
		cdc:          app.appCodec,
		BankKeeper:   app.BankKeeper,
		DollarKeeper: app.DollarKeeper,
		FTFKeeper:    app.FTFKeeper,
		IBCKeeper:    app.IBCKeeper,
	})
	if err != nil {
		return nil, err
	}
	app.SetAnteHandler(anteHandler)

	jesterClient := jester.NewClient(cast.ToString(appOpts.Get(jester.FlagGRPCAddress)))
	proposalHandler := NewProposalHandler(
		app.BaseApp, app.Mempool(), app.PreBlocker, app.txConfig,
		jesterClient, app.DollarKeeper, app.WormholeKeeper,
	)

	app.SetPrepareProposal(proposalHandler.PrepareProposal())
	app.SetPreBlocker(proposalHandler.PreBlocker())

	if err := app.RegisterUpgradeHandler(); err != nil {
		return nil, err
	}

	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	return app, nil
}

// InitAppForTestnet executes the necessary state transitions on the
// provided application in order to start an in-place testnet.
func InitAppForTestnet(app *App, pubKey crypto.PubKey, rawConsensusAddress bytes.HexBytes, rawOperatorAddress string, upgradeToTrigger string) *App {
	ctx := app.NewUncachedContext(true, cmtproto.Header{})

	// ===== Cosmos SDK: Staking  =====

	cmtPubKey := &ed25519.PubKey{Key: pubKey.Bytes()}
	operatorAddress := sdk.MustBech32ifyAddressBytes("noblevaloper", sdk.MustAccAddressFromBech32(rawOperatorAddress))
	validator, err := stakingtypes.NewValidator(
		operatorAddress,
		cmtPubKey,
		stakingtypes.Description{Moniker: "Testnet Validator"},
	)
	if err != nil {
		cmtos.Exit("failed to create new validator: " + err.Error())
	}

	validator.Status = stakingtypes.Bonded
	validator.Tokens = math.NewInt(1000000)
	validator.DelegatorShares = math.LegacyNewDec(1000000)

	stakingStore := ctx.KVStore(app.GetKey(stakingtypes.ModuleName))

	// Remove all validators from last validators store
	iterator, err := app.StakingKeeper.LastValidatorsIterator(ctx)
	if err != nil {
		cmtos.Exit("failed to create last validators iterator: " + err.Error())
	}
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from power store
	iterator, err = app.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		cmtos.Exit("failed to create validators power store iterator: " + err.Error())
	}
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Remove all validators from validators store
	iterator = storetypes.KVStorePrefixIterator(stakingStore, stakingtypes.ValidatorsKey)
	for ; iterator.Valid(); iterator.Next() {
		stakingStore.Delete(iterator.Key())
	}
	iterator.Close()

	// Add our validator to power and last validators store
	err = app.StakingKeeper.SetValidator(ctx, validator)
	if err != nil {
		cmtos.Exit("failed to set validator: " + err.Error())
	}
	err = app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	if err != nil {
		cmtos.Exit("failed to set validator by consensus address: " + err.Error())
	}
	err = app.StakingKeeper.SetValidatorByPowerIndex(ctx, validator)
	if err != nil {
		cmtos.Exit("failed to set validator by power index: " + err.Error())
	}
	valAddress, _ := sdk.ValAddressFromBech32(validator.GetOperator())
	err = app.StakingKeeper.SetLastValidatorPower(ctx, valAddress, 0)
	if err != nil {
		cmtos.Exit("failed to set last validator power: " + err.Error())
	}
	if err := app.StakingKeeper.Hooks().AfterValidatorCreated(ctx, valAddress); err != nil {
		cmtos.Exit("failed to execute after validator created hooks: " + err.Error())
	}

	// ===== Cosmos SDK: Slashing =====

	consensusAddress := sdk.ConsAddress(rawConsensusAddress.Bytes())
	newValidatorSigningInfo := slashingtypes.ValidatorSigningInfo{
		Address:     consensusAddress.String(),
		StartHeight: app.LastBlockHeight() - 1,
		Tombstoned:  false,
	}
	err = app.SlashingKeeper.SetValidatorSigningInfo(ctx, consensusAddress, newValidatorSigningInfo)
	if err != nil {
		cmtos.Exit("failed to set validator signing info: " + err.Error())
	}

	// ===== Cosmos SDK: Upgrade  =====

	if upgradeToTrigger != "" {
		upgradePlan := upgradetypes.Plan{
			Name:   upgradeToTrigger,
			Height: app.LastBlockHeight() + 10,
		}
		err = app.UpgradeKeeper.ScheduleUpgrade(ctx, upgradePlan)
		if err != nil {
			cmtos.Exit("failed to schedule upgrade: " + err.Error())
		}
	}

	return app
}

func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *App) SimulationManager() *module.SimulationManager {
	return nil
}

func (app *App) RegisterAPIRoutes(apiSvr *serverapi.Server, apiConfig serverconfig.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)

	if err := api.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}
}

//

func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	key, _ := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)
	return key
}

func (app *App) GetMemKey(memKey string) *storetypes.MemoryStoreKey {
	key, _ := app.UnsafeFindStoreKey(memKey).(*storetypes.MemoryStoreKey)
	return key
}

func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

func (app *App) RegisterUpgradeHandler() error {
	app.UpgradeKeeper.SetUpgradeHandler(
		upgrade.UpgradeName,
		upgrade.CreateUpgradeHandler(
			app.ModuleManager,
			app.Configurator(),
			app.Logger(),
			app.AccountKeeper.AddressCodec(),
			app.AuthorityKeeper,
			app.BankKeeper,
			app.IBCKeeper.ClientKeeper,
			app.DollarKeeper,
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		return fmt.Errorf("failed to read upgrade info from disk: %w", err)
	}
	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return nil
	}

	if upgradeInfo.Name == upgrade.UpgradeName {
		app.SetStoreLoader(upgrade.CreateStoreLoader(upgradeInfo.Height))
	}

	return nil
}

func (app *App) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}
