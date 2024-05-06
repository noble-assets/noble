package krypton

import (
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	authoritytypes "github.com/noble-assets/authority/x/authority/types"
)

func CreateStoreLoader(upgradeHeight int64) baseapp.StoreLoader {
	storeUpgrades := storetypes.StoreUpgrades{
		Added: []string{
			// SDK Modules
			consensustypes.StoreKey,
			crisistypes.StoreKey,
			grouptypes.StoreKey,
			// Noble Modules
			authoritytypes.ModuleName,
		},
	}

	return upgradetypes.UpgradeStoreLoader(upgradeHeight, &storeUpgrades)
}
