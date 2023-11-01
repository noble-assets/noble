package v4m0p0rc0

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Ensure that this upgrade is only run on Noble's testnet.
		if ctx.ChainID() != TestnetChainID {
			return vm, errors.New(fmt.Sprintf("%s upgrade not allowed to execute on %s chain", UpgradeName, ctx.ChainID()))
		}

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
