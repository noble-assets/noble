package upgrade

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	authoritytypes "github.com/noble-assets/authority/types"
)

func CreateStoreLoader(upgradeHeight int64) baseapp.StoreLoader {
	storeUpgrades := storetypes.StoreUpgrades{
		Added: []string{
			// Cosmos Modules
			consensustypes.StoreKey,
			// Noble Modules
			authoritytypes.ModuleName,
		},
	}

	return upgradetypes.UpgradeStoreLoader(upgradeHeight, &storeUpgrades)
}
