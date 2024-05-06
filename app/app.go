package app

import (
	_ "embed"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	authoritykeeper "github.com/noble-assets/authority/x/authority/keeper"
	forwardingkeeper "github.com/noble-assets/forwarding/v2/x/forwarding/keeper"

	_ "cosmossdk.io/x/evidence"                                        // import for side effects
	_ "cosmossdk.io/x/feegrant/module"                                 // import for side effects
	_ "cosmossdk.io/x/upgrade"                                         // import for side effects
	_ "github.com/circlefin/noble-cctp/x/cctp"                         // import for side effects
	_ "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory" // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/auth"                            // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"                    // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/authz/module"                    // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"                            // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/consensus"                       // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/crisis"                          // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/distribution"                    // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/group/module"                    // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/params"                          // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/slashing"                        // import for side effects
	_ "github.com/cosmos/cosmos-sdk/x/staking"                         // import for side effects
	_ "github.com/noble-assets/authority/x/authority"                  // import for side effects
)

var DefaultNodeHome string

//go:embed app.yaml
var AppConfigYAML []byte

var (
	_ runtime.AppI            = (*NobleApp)(nil)
	_ servertypes.Application = (*NobleApp)(nil)
)

// NobleApp ... TODO
type NobleApp struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// SDK Modules
	AccountKeeper      authkeeper.AccountKeeper
	AuthzKeeper        authzkeeper.Keeper
	BankKeeper         bankkeeper.Keeper
	ConsensusKeeper    consensuskeeper.Keeper
	CrisisKeeper       *crisiskeeper.Keeper
	DistributionKeeper distributionkeeper.Keeper
	EvidenceKeeper     evidencekeeper.Keeper
	FeeGrantKeeper     feegrantkeeper.Keeper
	GroupKeeper        groupkeeper.Keeper
	ParamsKeeper       paramskeeper.Keeper
	SlashingKeeper     slashingkeeper.Keeper
	StakingKeeper      *stakingkeeper.Keeper
	UpgradeKeeper      *upgradekeeper.Keeper
	// IBC Modules
	CapabilityKeeper     *capabilitykeeper.Keeper
	IBCKeeper            *ibckeeper.Keeper
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ICAHostKeeper        icahostkeeper.Keeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper
	TransferKeeper       transferkeeper.Keeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	PFMKeeper            *packetforwardkeeper.Keeper
	// Circle Modules
	FiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper
	CCTPKeeper             *cctpkeeper.Keeper
	// Noble Modules
	AuthorityKeeper  *authoritykeeper.Keeper
	ForwardingKeeper *forwardingkeeper.Keeper
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

// NewNobleApp returns a reference to an initialized NobleApp.
func NewNobleApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOptions servertypes.AppOptions,
	baseappOptions ...func(*baseapp.BaseApp),
) (*NobleApp, error) {
	var (
		app        = &NobleApp{}
		appBuilder *runtime.AppBuilder
	)

	if err := depinject.Inject(
		depinject.Configs(
			AppConfig(),
			depinject.Supply(
				logger,
				appOptions,
			),
		),
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		// SDK Modules
		&app.AccountKeeper,
		&app.AuthzKeeper,
		&app.BankKeeper,
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
		// Circle Modules
		&app.FiatTokenFactoryKeeper,
		&app.CCTPKeeper,
		// Noble Modules
		&app.AuthorityKeeper,
		&app.ForwardingKeeper,
	); err != nil {
		return nil, err
	}

	app.App = appBuilder.Build(db, traceStore, baseappOptions...)

	if err := app.RegisterIBCModules(); err != nil {
		panic(err)
	}

	// TODO(@john): Register ante handler.

	app.RegisterUpgradeHandlers()

	if err := app.RegisterStreamingServices(appOptions, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	return app, nil
}

// LegacyAmino implements the runtime.AppI interface.
func (app *NobleApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// SimulationManager implements the runtime.AppI interface.
func (app *NobleApp) SimulationManager() *module.SimulationManager {
	return nil
}

//

func (app *NobleApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	key, _ := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)
	return key
}

func (app *NobleApp) GetMemKey(memKey string) *storetypes.MemoryStoreKey {
	key, _ := app.UnsafeFindStoreKey(memKey).(*storetypes.MemoryStoreKey)
	return key
}

func (app *NobleApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

func (app *NobleApp) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}
