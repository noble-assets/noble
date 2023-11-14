package v4m1p0rc0

import (
	"errors"
	"fmt"

	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory"
	stabletokenfactorykeeper "github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/keeper"
	stabletokenfactorytypes "github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
	tarifftypes "github.com/strangelove-ventures/noble/v4/x/tariff/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	cdc *codec.LegacyAmino,
	accountKeeper authkeeper.AccountKeeper,
	bankKeeper bankkeeper.Keeper,
	fiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper,
	stableTokenFactoryKeeper *stabletokenfactorykeeper.Keeper,
	tariffSubspace paramstypes.Subspace,
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

		// Update x/tariff module parameters to include $USDLR fees.
		fiatMintingDenom := fiatTokenFactoryKeeper.GetMintingDenom(ctx)

		transferFees := []tarifftypes.TransferFee{
			{
				Bps:   sdk.OneInt(),
				Max:   sdk.NewInt(5_000_000),
				Denom: fiatMintingDenom.Denom,
			},
			{
				Bps:   sdk.OneInt(),
				Max:   sdk.NewInt(5_000_000),
				Denom: USDLRMetadata.Base,
			},
		}

		err := tariffSubspace.Update(ctx, tarifftypes.KeyTransferFees, cdc.MustMarshalJSON(transferFees))
		if err != nil {
			return vm, err
		}

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
