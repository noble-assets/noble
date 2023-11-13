package v4m1p0rc0

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory"
	stabletokenfactorykeeper "github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/keeper"
	stabletokenfactorytypes "github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	stableTokenFactoryKeeper *stabletokenfactorykeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Ensure that this upgrade is only run on Noble's testnet.
		if ctx.ChainID() != TestnetChainID {
			return vm, errors.New(fmt.Sprintf("%s upgrade not allowed to execute on %s chain", UpgradeName, ctx.ChainID()))
		}

		// Set metadata in the x/bank module for the $USDLR token.
		bankKeeper.SetDenomMetaData(ctx, USDLRMetadata)

		// Ensure that the account owned by Stable exists on chain.
		StableAccAddress := sdk.MustAccAddressFromBech32(StableAddress)
		if !accountKeeper.HasAccount(ctx, StableAccAddress) {
			// The Stable account doesn't exist, let's initialise it.
			account := accountKeeper.NewAccountWithAddress(ctx, StableAccAddress)
			accountKeeper.SetAccount(ctx, account)
		}

		// Configure permissions and roles for the x/stabletokenfactory module.
		genesis := stabletokenfactorytypes.GenesisState{
			Paused:       &stabletokenfactorytypes.Paused{Paused: false},
			MasterMinter: &stabletokenfactorytypes.MasterMinter{Address: StableAddress},
			Pauser:       &stabletokenfactorytypes.Pauser{Address: StableAddress},
			Blacklister:  &stabletokenfactorytypes.Blacklister{Address: StableAddress},
			Owner:        &stabletokenfactorytypes.Owner{Address: StableAddress},
			MintingDenom: &stabletokenfactorytypes.MintingDenom{Denom: USDLRMetadata.Base},
		}

		stabletokenfactory.InitGenesis(ctx, stableTokenFactoryKeeper, bankKeeper, genesis)
		vm[stabletokenfactorytypes.ModuleName] = stabletokenfactory.AppModule{}.ConsensusVersion()

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
