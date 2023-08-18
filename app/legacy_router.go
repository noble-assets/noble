package app

import (
	// Gov
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	// IBC Core
	ibc "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcTypes "github.com/cosmos/ibc-go/v7/modules/core/exported"
	// Params
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	// Upgrade
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (app *NobleApp) RegisterLegacyRouter() {
	router := v1beta1.NewRouter()
	var handler v1beta1.Handler

	handler = ibc.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)
	router.AddRoute(ibcTypes.RouterKey, handler)

	handler = params.NewParamChangeProposalHandler(app.ParamsKeeper)
	router.AddRoute(paramsTypes.RouterKey, handler)

	handler = upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)
	router.AddRoute(upgradeTypes.RouterKey, handler)

	app.AuthorityKeeper.SetLegacyRouter(router)
}
