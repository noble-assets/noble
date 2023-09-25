package argon2

import (
	"cosmossdk.io/math"
	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
		authority := paramauthoritykeeper.GetAuthority(ctx)

		denom := fiatTFKeeper.GetMintingDenom(ctx)

		cctpKeeper.SetOwner(ctx, authority)
		cctpKeeper.SetAttesterManager(ctx, "noble1rx6vk9ll2vglwkrrlf8a7cl86lfz55uj8deunm")
		cctpKeeper.SetPauser(ctx, "noble1hvm5pxssempk3jg0tgzugtsk85js42rze7cnxd")
		cctpKeeper.SetTokenController(ctx, "noble1hl7nlkt3vyjzk0c4ytfveemmykw8ectspapcd3")
		cctpKeeper.SetPerMessageBurnLimit(ctx, cctptypes.PerMessageBurnLimit{Denom: denom.Denom, Amount: math.NewInt(99999999)})
		cctpKeeper.SetBurningAndMintingPaused(ctx, cctptypes.BurningAndMintingPaused{Paused: false})
		cctpKeeper.SetSendingAndReceivingMessagesPaused(ctx, cctptypes.SendingAndReceivingMessagesPaused{Paused: false})
		cctpKeeper.SetMaxMessageBodySize(ctx, cctptypes.MaxMessageBodySize{Amount: 8000})
		cctpKeeper.SetSignatureThreshold(ctx, cctptypes.SignatureThreshold{Amount: 2})

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
