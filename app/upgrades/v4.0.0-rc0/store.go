package v4m0p0rc0

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateStoreLoader(upgradeHeight int64) baseapp.StoreLoader {
	storeUpgrades := storetypes.StoreUpgrades{
		Deleted: []string{"router"},
	}

	return upgradetypes.UpgradeStoreLoader(upgradeHeight, &storeUpgrades)
}
