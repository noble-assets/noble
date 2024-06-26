package krypton

import (
	"fmt"

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

		switch ctx.ChainID() {
		case TestnetChainID:
			auraKeeper.SetOwner(ctx, "noble1mxe0zwwdvjvn8dg2hnep55q4fc7sqmpud9qsqn")
			auraKeeper.SetBlocklistOwner(ctx, "noble1mxe0zwwdvjvn8dg2hnep55q4fc7sqmpud9qsqn")
		default:
			return vm, fmt.Errorf("%s upgrade not allowed to execute on %s chain", UpgradeName, ctx.ChainID())
		}

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
