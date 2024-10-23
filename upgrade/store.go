package upgrade

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	authoritytypes "github.com/noble-assets/authority/types"
	globalfeetypes "github.com/noble-assets/globalfee/types"
)

func CreateStoreLoader(upgradeHeight int64) baseapp.StoreLoader {
	storeUpgrades := storetypes.StoreUpgrades{
		Added: []string{
			// Cosmos Modules
			consensustypes.StoreKey,
			crisistypes.StoreKey,
			// Noble Modules
			authoritytypes.ModuleName,
			globalfeetypes.ModuleName,
		},
	}

	return upgradetypes.UpgradeStoreLoader(upgradeHeight, &storeUpgrades)
}
