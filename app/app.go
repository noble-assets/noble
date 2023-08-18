package app

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/depinject"

	cmtDb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	// Auth
	"github.com/cosmos/cosmos-sdk/x/auth"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side effects
	// Authority
	"github.com/noble-assets/paramauthority/x/authority"
	authorityKeeper "github.com/noble-assets/paramauthority/x/authority/keeper"
	// Authz
	authzKeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	// Bank
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	// Capability
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilityKeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	// Consensus
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusKeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	// Crisis
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisisKeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	// Distribution
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributionKeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	// Evidence
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidenceKeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	// FeeGrant
	feeGrantKeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feeGrant "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	// FiatTokenFactory
	fiatTokenFactory "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	fiatTokenFactoryKeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	// GenUtil
	genUtil "github.com/cosmos/cosmos-sdk/x/genutil"
	genUtilTypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	// GlobalFee
	globalFee "github.com/strangelove-ventures/noble/x/globalfee"
	globalFeeKeeper "github.com/strangelove-ventures/noble/x/globalfee/keeper"
	// Group
	groupKeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	group "github.com/cosmos/cosmos-sdk/x/group/module"
	// IBC Client
	ibcClientSolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibcClientTendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	// IBC Core
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcKeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	// IBC Fee
	ibcFee "github.com/cosmos/ibc-go/v7/modules/apps/29-fee"
	ibcFeeKeeper "github.com/cosmos/ibc-go/v7/modules/apps/29-fee/keeper"
	// IBC Transfer
	ibcTransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibcTransferKeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	// ICA
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	// ICA Controller
	icaControllerKeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	// ICA Host
	icaHostKeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	// Params
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsKeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
	// PFM
	pfm "github.com/strangelove-ventures/packet-forward-middleware/v7/router"
	pfmKeeper "github.com/strangelove-ventures/packet-forward-middleware/v7/router/keeper"
	// Slashing
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingKeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	// Staking
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingKeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	// Tariff
	"github.com/strangelove-ventures/noble/x/tariff"
	tariffKeeper "github.com/strangelove-ventures/noble/x/tariff/keeper"
	// TokenFactory
	tokenFactory "github.com/strangelove-ventures/noble/x/tokenfactory"
	tokenFactoryKeeper "github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	// Upgrade
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeKeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	// Vesting
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
)

var (
	DefaultNodeHome string

	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authz.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		consensus.AppModuleBasic{},
		crisis.AppModuleBasic{},
		distribution.AppModuleBasic{},
		evidence.AppModuleBasic{},
		feeGrant.AppModuleBasic{},
		genUtil.NewAppModuleBasic(genUtilTypes.DefaultMessageValidator),
		group.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		staking.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		vesting.AppModuleBasic{},

		ibc.AppModuleBasic{},
		ibcClientSolomachine.AppModuleBasic{},
		ibcClientTendermint.AppModuleBasic{},
		ibcFee.AppModuleBasic{},
		ibcTransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		pfm.AppModuleBasic{},

		authority.AppModuleBasic{},
		fiatTokenFactory.AppModuleBasic{},
		globalFee.AppModuleBasic{},
		tariff.AppModuleBasic{},
		tokenFactory.AppModuleBasic{},
	)
)

var (
	_ runtime.AppI            = (*NobleApp)(nil)
	_ serverTypes.Application = (*NobleApp)(nil)
)

type NobleApp struct {
	*runtime.App

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codecTypes.InterfaceRegistry

	// Cosmos SDK Keepers
	AccountKeeper      authKeeper.AccountKeeper
	AuthzKeeper        authzKeeper.Keeper
	BankKeeper         bankKeeper.Keeper
	CapabilityKeeper   *capabilityKeeper.Keeper
	ConsensusKeeper    consensusKeeper.Keeper
	CrisisKeeper       *crisisKeeper.Keeper
	DistributionKeeper distributionKeeper.Keeper
	EvidenceKeeper     evidenceKeeper.Keeper
	FeeGrantKeeper     feeGrantKeeper.Keeper
	GroupKeeper        groupKeeper.Keeper
	ParamsKeeper       paramsKeeper.Keeper
	SlashingKeeper     slashingKeeper.Keeper
	StakingKeeper      *stakingKeeper.Keeper
	UpgradeKeeper      *upgradeKeeper.Keeper

	// IBC Keepers
	IBCKeeper           *ibcKeeper.Keeper
	IBCFeeKeeper        ibcFeeKeeper.Keeper
	IBCTransferKeeper   ibcTransferKeeper.Keeper
	ICAControllerKeeper icaControllerKeeper.Keeper
	ICAHostKeeper       icaHostKeeper.Keeper
	PFMKeeper           *pfmKeeper.Keeper

	// Custom Keepers
	AuthorityKeeper        *authorityKeeper.Keeper
	FiatTokenFactoryKeeper *fiatTokenFactoryKeeper.Keeper
	GlobalFeeKeeper        *globalFeeKeeper.Keeper
	TariffKeeper           *tariffKeeper.Keeper
	TokenFactoryKeeper     *tokenFactoryKeeper.Keeper

	// Scoped Keepers (for IBC)
	ScopedIBCKeeper           capabilityKeeper.ScopedKeeper
	ScopedIBCTransferKeeper   capabilityKeeper.ScopedKeeper
	ScopedICAControllerKeeper capabilityKeeper.ScopedKeeper
	ScopedICAHostKeeper       capabilityKeeper.ScopedKeeper
	ScopedPFMKeeper           capabilityKeeper.ScopedKeeper
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".noble")
}

func NewNobleApp(
	logger log.Logger,
	db cmtDb.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts serverTypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *NobleApp {
	var (
		app        = &NobleApp{}
		appBuilder *runtime.AppBuilder

		appConfig = depinject.Configs(
			AppConfig,
			depinject.Supply(appOpts),
		)
	)

	if err := depinject.Inject(appConfig,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.AuthzKeeper,
		&app.BankKeeper,
		&app.CapabilityKeeper,
		&app.ConsensusKeeper,
		&app.CrisisKeeper,
		&app.DistributionKeeper,
		&app.EvidenceKeeper,
		&app.FeeGrantKeeper,
		&app.GroupKeeper,
		&app.ParamsKeeper,
		&app.SlashingKeeper,
		&app.StakingKeeper,
		&app.UpgradeKeeper,

		&app.AuthorityKeeper,
		&app.FiatTokenFactoryKeeper,
		&app.GlobalFeeKeeper,
		&app.TariffKeeper,
		&app.TokenFactoryKeeper,
	); err != nil {
		panic(err)
	}

	app.App = appBuilder.Build(logger, db, traceStore, baseAppOptions...)

	// Registers all modules that don't use App Wiring (e.g. IBC).
	app.RegisterLegacyModules()
	// Registers all proposals handlers that are using v1beta1 governance.
	app.RegisterLegacyRouter()

	if err := app.Load(loadLatest); err != nil {
		panic(err)
	}

	return app
}

func (app *NobleApp) GetSubspace(moduleName string) paramsTypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// ------------------------------- runtime.AppI --------------------------------

func (app *NobleApp) AppCodec() codec.Codec {
	return app.appCodec
}

func (app *NobleApp) ExportAppStateAndValidators(
	_ bool, _ []string, _ []string,
) (serverTypes.ExportedApp, error) {
	panic("UNIMPLEMENTED")
}

func (app *NobleApp) InterfaceRegistry() codecTypes.InterfaceRegistry {
	return app.interfaceRegistry
}

func (app *NobleApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

func (app *NobleApp) SimulationManager() *module.SimulationManager {
	panic("UNIMPLEMENTED")
}

func (app *NobleApp) TxConfig() client.TxConfig {
	return app.txConfig
}
