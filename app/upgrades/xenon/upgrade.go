package xenon

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	halokeeper "github.com/noble-assets/halo/x/halo/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	haloKeeper *halokeeper.Keeper,
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
		case MainnetChainID:
			haloKeeper.SetOwner(ctx, "")             // TODO
			haloKeeper.SetAggregatorOwner(ctx, "")   // TODO
			haloKeeper.SetEntitlementsOwner(ctx, "") // TODO
		default:
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, ctx.ChainID())
		}

		bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
			Description: "Hashnote US Yield Coin",
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
			Name:    "Hashnote US Yield Coin",
			Symbol:  "USYC",
		})

		return vm, nil
	}
}
