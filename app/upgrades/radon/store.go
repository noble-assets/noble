package radon

import (
	// GlobalFee
	globalFeeTypes "github.com/strangelove-ventures/noble/x/globalfee/types"
	// Tariff
	tariffTypes "github.com/strangelove-ventures/noble/x/tariff/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	// Upgrade
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func CreateStoreLoader(upgradeHeight int64) baseapp.StoreLoader {
	storeUpgrades := storeTypes.StoreUpgrades{
		Added: []string{
			globalFeeTypes.ModuleName, tariffTypes.ModuleName,
		},
	}

	return upgradeTypes.UpgradeStoreLoader(upgradeHeight, &storeUpgrades)
}
