package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/noble-assets/noble/v5/app/upgrades/krypton"
)

func (app *NobleApp) RegisterUpgradeHandlers() {
	app.UpgradeKeeper.SetUpgradeHandler(
		krypton.UpgradeName,
		krypton.CreateUpgradeHandler(
			app.ModuleManager,
			app.Configurator(),
			app.appCodec,
			app.Logger(),
			app.GetKey(capabilitytypes.StoreKey),
			app.AccountKeeper,
			app.AuthorityKeeper,
			app.BankKeeper,
			app.CapabilityKeeper,
			app.IBCKeeper.ClientKeeper,
			app.ConsensusKeeper,
			app.ParamsKeeper,
			app.StakingKeeper,
		),
	)

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Errorf("failed to read upgrade info from disk: %w", err))
	}
	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	var storeLoader baseapp.StoreLoader
	switch upgradeInfo.Name {
	case krypton.UpgradeName:
		storeLoader = krypton.CreateStoreLoader(upgradeInfo.Height)
	}
	if storeLoader != nil {
		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(storeLoader)
	}
}
