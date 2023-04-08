package radon

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	paramauthoritykeeper "github.com/strangelove-ventures/paramauthority/x/params/keeper"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	globalfeetypes "github.com/strangelove-ventures/noble/x/globalfee/types"

	tarifftypes "github.com/strangelove-ventures/noble/x/tariff/types"
)

func CreateRadonUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	paramauthoritykeeper paramauthoritykeeper.Keeper,
	fiatTFKeeper *fiattokenfactorykeeper.Keeper,

) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		// New modules run AFTER the migrations, so to set the correct params after the default
		// becasuse RunMigrations runs `InitGenesis` on new modules`.
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		// -- globalfee params --
		minGasPrices := sdk.DecCoins{
			sdk.NewDecCoinFromDec("udrachma", sdk.NewDecWithPrec(1, 2)),
		}
		globlaFeeParamsSubspace, ok := paramauthoritykeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}
		globlaFeeParamsSubspace.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)
		// -- --

		// -- tariff params --
		tariffParamsSubspace, ok := paramauthoritykeeper.GetSubspace(tarifftypes.ModuleName)
		if !ok {
			panic("tariff params subspace not found")
		}
		paramAuth := paramauthoritykeeper.GetAuthority(ctx)
		distributionEntities := []tarifftypes.DistributionEntity{
			{
				Address: paramAuth,
				Share:   sdk.NewDec(1),
			},
		}
		feeDenom := fiatTFKeeper.GetMintingDenom(ctx)
		tariffParams := tarifftypes.Params{
			Share:                sdk.NewDecWithPrec(8, 1),
			DistributionEntities: distributionEntities,
			TransferFeeBps:       sdk.OneInt(),
			TransferFeeMax:       sdk.NewInt(5000000),
			TransferFeeDenom:     feeDenom.Denom,
		}
		tariffParamsSubspace.SetParamSet(ctx, &tariffParams)
		// -- --

		return versionMap, err
	}
}
