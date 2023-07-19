package argon

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	cctpkeeper "github.com/strangelove-ventures/noble/x/cctp/keeper"
	cctptypes "github.com/strangelove-ventures/noble/x/cctp/types"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	paramauthoritykeeper "github.com/strangelove-ventures/paramauthority/x/params/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	fiatTFKeeper *fiattokenfactorykeeper.Keeper,
	paramauthoritykeeper paramauthoritykeeper.Keeper,
	cctpKeeper *cctpkeeper.Keeper,
) upgradeTypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradeTypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		var authority string
		if ctx.ChainID() == TestnetChainID {
			authority = paramauthoritykeeper.GetAuthority(ctx)
		} else {
			owner, ok := fiatTFKeeper.GetOwner(ctx)
			if !ok {
				return nil, fmt.Errorf("fiat token factory owner not found")
			}

			authority = owner.Address
		}

		cctpKeeper.SetAuthority(ctx, cctptypes.Authority{Address: authority})
		cctpKeeper.SetPerMessageBurnLimit(ctx, cctptypes.PerMessageBurnLimit{Amount: 99999999})
		cctpKeeper.SetBurningAndMintingPaused(ctx, cctptypes.BurningAndMintingPaused{Paused: false})
		cctpKeeper.SetSendingAndReceivingMessagesPaused(ctx, cctptypes.SendingAndReceivingMessagesPaused{Paused: false})
		cctpKeeper.SetMaxMessageBodySize(ctx, cctptypes.MaxMessageBodySize{Amount: 8000})
		cctpKeeper.SetSignatureThreshold(ctx, cctptypes.SignatureThreshold{Amount: 2})

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
