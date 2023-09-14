package v3m1p0

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateV3M1P0UpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
) upgradeTypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradeTypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
