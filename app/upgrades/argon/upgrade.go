package argon

import (
	"cosmossdk.io/math"
	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cctpKeeper *cctpkeeper.Keeper,
	fiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, nil
		}

		cctpKeeper.SetOwner(ctx, "noble127de05h6z3a3rh5jf0rjepa48zpgxtesfywgtf")
		cctpKeeper.SetAttesterManager(ctx, "noble1ak4d4dsrx5ec37h3qpsm8x6kg39xy0d0l8ptdq")
		cctpKeeper.SetPauser(ctx, "noble1cnl6q0c7g3aq8fjgeh9ygy5p2gv83kxqp4pfw4")
		cctpKeeper.SetTokenController(ctx, "noble1ye45j5c5gks2r68z6s8k9aehma372r927nuze4")

		denom := fiatTokenFactoryKeeper.GetMintingDenom(ctx)
		cctpKeeper.SetPerMessageBurnLimit(ctx, cctptypes.PerMessageBurnLimit{
			Denom:  denom.Denom,
			Amount: math.NewInt(1_000_000_000_000),
		})

		cctpKeeper.SetBurningAndMintingPaused(ctx, cctptypes.BurningAndMintingPaused{Paused: false})
		cctpKeeper.SetSendingAndReceivingMessagesPaused(ctx, cctptypes.SendingAndReceivingMessagesPaused{Paused: false})

		cctpKeeper.SetMaxMessageBodySize(ctx, cctptypes.MaxMessageBodySize{Amount: 8192})
		cctpKeeper.SetSignatureThreshold(ctx, cctptypes.SignatureThreshold{Amount: 2})

		return vm, nil
	}
}
