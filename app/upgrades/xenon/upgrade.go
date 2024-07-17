package xenon

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	florinkeeper "github.com/noble-assets/florin/x/florin/keeper"
	halokeeper "github.com/noble-assets/halo/x/halo/keeper"
)

var (
	// TODO: Verify denom metadata

	eureMetadata = banktypes.Metadata{
		Description: "Regulated Euro Stablecoin",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "aeure",
				Exponent: 0,
				Aliases:  []string{"attoeure"},
			},
			{
				Denom:    "eure",
				Exponent: 18,
			},
		},
		Base:    "aeure",
		Display: "eure",
		Name:    "Euro Stablecoin",
		Symbol:  "EURe",
	}

	usycMetadata = banktypes.Metadata{
		Description: "",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uusyc",
				Exponent: 0,
				Aliases:  []string{"microusyc"},
			},
			{
				Denom:    "usyc",
				Exponent: 6,
			},
		},
		Base:    "uusyc",
		Display: "usyc",
		Name:    "",
		Symbol:  "USYC",
	}
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	haloKeeper *halokeeper.Keeper,
	florinKeeper *florinkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		switch ctx.ChainID() {
		case TestnetChainID:
			haloKeeper.SetOwner(ctx, "noble1u0nahk4wltsp89tpce4cyayd63a69dhpkfq9wq")
			haloKeeper.SetAggregatorOwner(ctx, "noble1u0nahk4wltsp89tpce4cyayd63a69dhpkfq9wq")
			haloKeeper.SetEntitlementsOwner(ctx, "noble1u0nahk4wltsp89tpce4cyayd63a69dhpkfq9wq")

			florinKeeper.SetOwner(ctx, "noble1tv9u97jln0k3anpzhahkeahh66u74dug302pyn")
			florinKeeper.SetBlacklistOwner(ctx, "noble1tv9u97jln0k3anpzhahkeahh66u74dug302pyn")
		case MainnetChainID:
			haloKeeper.SetOwner(ctx, "")             // TODO
			haloKeeper.SetAggregatorOwner(ctx, "")   // TODO
			haloKeeper.SetEntitlementsOwner(ctx, "") // TODO

			florinKeeper.SetOwner(ctx, "")          // TODO
			florinKeeper.SetBlacklistOwner(ctx, "") // TODO
		default:
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, ctx.ChainID())
		}

		bankKeeper.SetDenomMetaData(ctx, eureMetadata)
		bankKeeper.SetDenomMetaData(ctx, usycMetadata)

		return vm, nil
	}
}
