package neon

import (
	"github.com/strangelove-ventures/noble/x/fiattokenfactory"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func initialFiatTokenFactoryState() fiattokenfactorytypes.GenesisState {
	s := fiattokenfactorytypes.DefaultGenesis()

	s.Owner = &fiattokenfactorytypes.Owner{
		Address: "noble1ljd4ywg3a5rrnxgq2c98pzcjq99f4kl764tmjv",
	}

	s.MasterMinter = &fiattokenfactorytypes.MasterMinter{
		Address: "noble1qty27zcvl2sgzzklz9syl0jc978ufm2e8mpufq",
	}

	s.Blacklister = &fiattokenfactorytypes.Blacklister{
		Address: "noble1jvx5x7pnjsaw80uc6j2fmv0he4kymg4dva0gfx",
	}

	s.Pauser = &fiattokenfactorytypes.Pauser{
		Address: "noble1szdzqxvq99vrpdys66nlp3q3794yuvvkp45mxj",
	}

	s.MinterControllerList = []fiattokenfactorytypes.MinterController{
		{
			Controller: "noble1xjz2j7y62us6famtq7fyfnenwv0k5yzhmsgaqt",
			Minter:     "noble18hn9z6wggf665vnqnvjs084tj84ysjjhq0y555",
		},
		{
			Controller: "noble1fetue2t0t6qxj579986425n4m2rhpp6hxtm7pq",
			Minter:     "noble10yyx9vs73gg6v46lcxl4hp2cgw95j4tjr9dk9w",
		},
		{
			Controller: "noble1v4t7awfpx6vw4mf9lyalu8qjf3sm8nfutl090f",
			Minter:     "noble1aq82vs8vwt0yqxljqcv36x5e6gvk775dcgs22u",
		},
		{
			Controller: "noble1uxckxfngckvg8jkjfk3yl9dwknvgkvsdtututz",
			Minter:     "noble1asdm30ncj4yzmgxdpfcuq0m4mxukx7xde9nuuy",
		},
	}

	s.MintingDenom = &fiattokenfactorytypes.MintingDenom{
		Denom: "uusdc",
	}

	return *s
}

var denomMetadataUsdc = banktypes.Metadata{
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
