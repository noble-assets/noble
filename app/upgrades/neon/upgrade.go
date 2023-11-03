package neon

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
)

func initialFiatTokenFactoryState() fiattokenfactorytypes.GenesisState {
	s := fiattokenfactorytypes.DefaultGenesis()

	s.Owner = &fiattokenfactorytypes.Owner{
		// TEST-MN: leg stove oblige forest occur range jar observe ahead morning street forward amazing negative digital syrup bar doctor fortune purpose buddy quote laptop civil
		Address: "noble1mp0z2yrmq86l53kr8gjzml3ctyf33ma969zjt6",
	}

	s.MasterMinter = &fiattokenfactorytypes.MasterMinter{
		Address: "noble1fknmpexguqlwu0pvgjgcktw525yy0r5r504mnp",
	}

	s.Blacklister = &fiattokenfactorytypes.Blacklister{
		Address: "noble1nklvu0y324jult8h3ymtn3lg5064k8jdwmzgd0",
	}

	s.Pauser = &fiattokenfactorytypes.Pauser{
		Address: "noble1dug3wwc995jvmhjrx9k34tvfrzprvfuedu49y5",
	}

	s.MinterControllerList = []fiattokenfactorytypes.MinterController{
		{
			Controller: "noble1rq6m2g3hqflk6zm3pmf6h49ufjm9w9r9ue32yr",
			Minter:     "noble1n35s7ytfyqrmhkjjwd06ltztjgxyyrutwlrncc",
		},
		{
			Controller: "noble1f7ylpwvyf4cuy9t026jr56gnfykmgeau0rger2",
			Minter:     "noble1gezp6maa6wjle5weqqjfy9s58gce4m3arzgjty",
		},
		{
			Controller: "noble1rn99mk9fxvqmsmfwe3y4spzmgd9v4ae2k79vup",
			Minter:     "noble1r9w4c9nws79krvdqx58k9jpt8sng68rhqmdtqx",
		},
		{
			Controller: "noble1hftnfd8tp6zn4marfvvkhldyk0jpr2ynzp4xey",
			Minter:     "noble1yjlapww37ryydskg5x6tpfugp0n8wasnzshlyq",
		},
	}

	s.MintingDenom = &fiattokenfactorytypes.MintingDenom{
		Denom: "uusdc",
	}

	return *s
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
	fiatTFKeeper fiattokenfactorykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	accountKeeper authkeeper.AccountKeeper,

) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		logger := ctx.Logger().With("upgrade", UpgradeName)

		// NOTE: denomMetadata must be set before setting the minting denom
		logger.Debug("adding usdc to bank denom metadata")
		bankKeeper.SetDenomMetaData(ctx, denomMetadataUsdc)

		logger.Debug("setting fiat-tokenfactory params")
		fiatTokenFactoryParams := initialFiatTokenFactoryState()
		fiattokenfactory.InitGenesis(ctx, &fiatTFKeeper, bankKeeper, fiatTokenFactoryParams)

		logger.Debug("adding fiat-tokenfactory accounts to account keeper")
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.Owner.Address)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MasterMinter.Address)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.Blacklister.Address)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.Pauser.Address)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[0].Controller)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[0].Minter)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[1].Controller)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[1].Minter)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[2].Controller)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[2].Minter)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[3].Controller)))
		accountKeeper.SetAccount(ctx, accountKeeper.NewAccountWithAddress(ctx, sdk.MustAccAddressFromBech32(fiatTokenFactoryParams.MinterControllerList[3].Minter)))

		return mm.RunMigrations(ctx, cfg, vm)
	}
}
