package argon

import (
	"cosmossdk.io/math"
	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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

		cctpKeeper.SetOwner(ctx, "noble1ye45j5c5gks2r68z6s8k9aehma372r927nuze4")
		cctpKeeper.SetAttesterManager(ctx, "noble1ak4d4dsrx5ec37h3qpsm8x6kg39xy0d0l8ptdq")
		cctpKeeper.SetPauser(ctx, "noble1cnl6q0c7g3aq8fjgeh9ygy5p2gv83kxqp4pfw4")
		cctpKeeper.SetTokenController(ctx, "noble1ye45j5c5gks2r68z6s8k9aehma372r927nuze4")

		// The below attesters are obtained from Circle's Iris API.
		// https://iris-api.circle.com/v1/publicKeys
		cctpKeeper.SetAttester(ctx, cctptypes.Attester{Attester: "0x04702317a335170cb26fef7577eeb5009451f72aca4ac5c03e330f68dd6a0d73728d2047346f216d9f3abc0337e77ed5e3b4995cd60cfa92f523faa29bce34e08b"})
		cctpKeeper.SetAttester(ctx, cctptypes.Attester{Attester: "0x0414f25da528fa94f46f081d4be46bcee81cb873297072cfcff0d60737e649d52158bebd0ed79f87959f152e0bb737de80574f79828b21c2b7e8a30b10fd6a56c5"})

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
