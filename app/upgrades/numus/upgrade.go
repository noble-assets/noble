package numus

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	florinkeeper "github.com/monerium/module-noble/x/florin/keeper"
	paramskeeper "github.com/strangelove-ventures/paramauthority/x/params/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	bankKeeper bankkeeper.Keeper,
	florinKeeper *florinkeeper.Keeper,
	paramsKeeper paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		authority := paramsKeeper.GetAuthority(ctx)
		florinKeeper.SetAuthority(ctx, authority)

		switch ctx.ChainID() {
		case TestnetChainID:
			florinKeeper.SetOwner(ctx, "ueure", "noble1tv9u97jln0k3anpzhahkeahh66u74dug302pyn")
			florinKeeper.SetBlacklistOwner(ctx, "noble1tv9u97jln0k3anpzhahkeahh66u74dug302pyn")
		case MainnetChainID:
			florinKeeper.SetOwner(ctx, "ueure", "noble1ya7ggnwv78qcnkv89lte30yge54ztzst3usgmw")
			florinKeeper.SetBlacklistOwner(ctx, "noble1ya7ggnwv78qcnkv89lte30yge54ztzst3usgmw")
		default:
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, ctx.ChainID())
		}

		bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
			Description: "Monerium EUR emoney",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "ueure",
					Exponent: 0,
					Aliases:  []string{"microeure"},
				},
				{
					Denom:    "eure",
					Exponent: 6,
				},
			},
			Base:    "ueure",
			Display: "eure",
			Name:    "Monerium EUR emoney",
			Symbol:  "EURe",
		})

		return vm, nil
	}
}
