package radon

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	globalfeetypes "github.com/strangelove-ventures/noble/x/globalfee/types"

	// paramauthoritykeeper "github.com/strangelove-ventures/paramauthority/x/params/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
)

func CreateRadonUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	// paramsKeeper paramauthoritykeeper.Keeper,
	paramsKeeper paramskeeper.Keeper,

) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// New modules run AFTER the migrations, so to set the correct params after the default
		// becasuse RunMigrations runs `InitGenesis` on the new module`.
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))

		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		minGasPrices := sdk.DecCoins{
			sdk.NewDecCoinFromDec("uusdc", sdk.NewDecWithPrec(1, 2)),
		}

		s, ok := paramsKeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}

		s.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)

		return versionMap, err
	}
}
