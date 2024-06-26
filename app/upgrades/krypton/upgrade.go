package krypton

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	aurakeeper "github.com/noble-assets/aura/x/aura/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	auraKeeper *aurakeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		auraKeeper.SetOwner(ctx, "")          // TODO
		auraKeeper.SetBlocklistOwner(ctx, "") // TODO

		bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
			Description: "Ondo US Dollar Yield",
			DenomUnits: []*banktypes.DenomUnit{
				{
					Denom:    "ausdy",
					Exponent: 0,
					Aliases:  []string{"attousdy"},
				},
				{
					Denom:    "usdy",
					Exponent: 18,
				},
			},
			Base:    "ausdy",
			Display: "usdy",
			Name:    "Ondo US Dollar Yield",
			Symbol:  "USDY",
		})

		return vm, nil
	}
}
