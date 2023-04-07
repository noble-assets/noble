package radon

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	globalfeetypes "github.com/strangelove-ventures/noble/x/globalfee/types"
)

func CreateRadonUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	paramsKeeper paramskeeper.Keeper,

) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// New modules run AFTER the migrations, so to set the correct params after the default
		// becasuse RunMigrations runs `InitGenesis` on the new module`.
		logger.Info(fmt.Sprintf("pre migrate version map: %v", vm))

		distributionParamSubspace, ok := paramsKeeper.GetSubspace(distrtypes.ModuleName)
		if !ok {
			panic("distribution params subspace not found")
		}

		logger.Info("setting distribution params...")
		distributionParamSubspace.Set(ctx, distrtypes.ParamStoreKeyCommunityTax, sdk.NewDec(0))
		distributionParamSubspace.Set(ctx, distrtypes.ParamStoreKeyBaseProposerReward, sdk.NewDec(0))
		distributionParamSubspace.Set(ctx, distrtypes.ParamStoreKeyBonusProposerReward, sdk.NewDec(0))
		distributionParamSubspace.Set(ctx, distrtypes.ParamStoreKeyWithdrawAddrEnabled, true)
		logger.Info("distribution params set")

		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		logger.Info(fmt.Sprintf("post migrate version map: %v", versionMap))

		minGasPrices := sdk.DecCoins{
			sdk.NewDecCoinFromDec("uusdc", sdk.NewDecWithPrec(1, 2)),
		}
		globlaFeeParamSubspace, ok := paramsKeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}

		globlaFeeParamSubspace.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)

		return versionMap, err
	}
}
