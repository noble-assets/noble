package argon3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	routerkeeper "github.com/strangelove-ventures/noble-router/x/router/keeper"
	routertypes "github.com/strangelove-ventures/noble-router/x/router/types"
	paramauthoritykeeper "github.com/strangelove-ventures/paramauthority/x/params/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	paramauthoritykeeper paramauthoritykeeper.Keeper,
	routerKeeper *routerkeeper.Keeper,
) upgradeTypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradeTypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		authority := paramauthoritykeeper.GetAuthority(ctx)
		routerKeeper.SetOwner(ctx, authority)

		routerKeeper.SetParams(ctx, routertypes.DefaultParams())

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
