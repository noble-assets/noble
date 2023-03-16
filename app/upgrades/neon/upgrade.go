package neon

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	circletokenfactorykeeper "github.com/strangelove-ventures/noble/x/circletokenfactory/keeper"
	circletokenfactorytypes "github.com/strangelove-ventures/noble/x/circletokenfactory/types"
)

var circleTokenFactoryParams = circletokenfactorytypes.GenesisState{
	Owner: &circletokenfactorytypes.Owner{
		Address: "noble10908tarhjl6zzlm4k6u9hkax48qeutk28tj4sp",
	},
	MintingDenom: &circletokenfactorytypes.MintingDenom{
		Denom: "uusdc",
	},
}

var (
	denomMetadataUsdc = banktypes.Metadata{

		Description: "USD Coin",
		Name:        "usdc",
		Base:        "uusdc",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom: "uusdc",
				Aliases: []string{
					"microusdc",
				},
				Exponent: 0,
			},
			{
				Denom:    "usdc",
				Exponent: 6,
			},
		},
	}
)

func CreateNeonUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	circletfkeeper circletokenfactorykeeper.Keeper,
	bankkeeper bankkeeper.Keeper,

) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// NOTE: denomMetadata must be set before setting the minting denom
		logger.Debug("adding usdc to bank denom metadata")
		bankkeeper.SetDenomMetaData(ctx, denomMetadataUsdc)

		logger.Debug("setting circle-tokenfactory params")
		circletfkeeper.SetParams(ctx, circletokenfactorytypes.DefaultParams())
		circletfkeeper.SetOwner(ctx, *circleTokenFactoryParams.Owner)
		circletfkeeper.SetMintingDenom(ctx, *circleTokenFactoryParams.MintingDenom)

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
