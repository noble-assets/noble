package radon

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	paramauthoritykeeper "github.com/strangelove-ventures/paramauthority/x/params/keeper"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	globalfeetypes "github.com/strangelove-ventures/noble/x/globalfee/types"

	tarifftypes "github.com/strangelove-ventures/noble/x/tariff/types"
)

func CreateRadonUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	paramauthoritykeeper paramauthoritykeeper.Keeper,
	fiatTFKeeper *fiattokenfactorykeeper.Keeper,

) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		// New modules run AFTER the migrations, so to set the correct params after the default
		// becasuse RunMigrations runs `InitGenesis` on new modules`.
		versionMap, err := mm.RunMigrations(ctx, cfg, vm)

		feeDenom := fiatTFKeeper.GetMintingDenom(ctx)

		// -- globalfee params --
		globalFeeParams := globalfeetypes.Params{
			MinimumGasPrices: sdk.DecCoins{
				sdk.NewDecCoinFromDec(feeDenom.Denom, sdk.NewDec(0)),
			},
			BypassMinFeeMsgTypes: []string{
				"/ibc.core.client.v1.MsgUpdateClient",
				"/ibc.core.channel.v1.MsgRecvPacket",
				"/ibc.core.channel.v1.MsgAcknowledgement",
				"/ibc.applications.transfer.v1.MsgTransfer",
				"/ibc.core.channel.v1.MsgTimeout",
				"/ibc.core.channel.v1.MsgTimeoutOnClose",
				"/cosmos.params.v1beta1.MsgUpdateParams",
				"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
				"/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
				"/noble.fiattokenfactory.MsgUpdateMasterMinter",
				"/noble.fiattokenfactory.MsgUpdatePauser",
				"/noble.fiattokenfactory.MsgUpdateBlacklister",
				"/noble.fiattokenfactory.MsgUpdateOwner",
				"/noble.fiattokenfactory.MsgAcceptOwner",
				"/noble.fiattokenfactory.MsgConfigureMinter",
				"/noble.fiattokenfactory.MsgRemoveMinter",
				"/noble.fiattokenfactory.MsgMint",
				"/noble.fiattokenfactory.MsgBurn",
				"/noble.fiattokenfactory.MsgBlacklist",
				"/noble.fiattokenfactory.MsgUnblacklist",
				"/noble.fiattokenfactory.MsgPause",
				"/noble.fiattokenfactory.MsgUnpause",
				"/noble.fiattokenfactory.MsgConfigureMinterController",
				"/noble.fiattokenfactory.MsgRemoveMinterController",
				"/noble.tokenfactory.MsgUpdatePauser",
				"/noble.tokenfactory.MsgUpdateBlacklister",
				"/noble.tokenfactory.MsgUpdateOwner",
				"/noble.tokenfactory.MsgAcceptOwner",
				"/noble.tokenfactory.MsgConfigureMinter",
				"/noble.tokenfactory.MsgRemoveMinter",
				"/noble.tokenfactory.MsgMint",
				"/noble.tokenfactory.MsgBurn",
				"/noble.tokenfactory.MsgBlacklist",
				"/noble.tokenfactory.MsgUnblacklist",
				"/noble.tokenfactory.MsgPause",
				"/noble.tokenfactory.MsgUnpause",
				"/noble.tokenfactory.MsgConfigureMinterController",
				"/noble.tokenfactory.MsgRemoveMinterController",
			},
		}
		globlaFeeParamsSubspace, ok := paramauthoritykeeper.GetSubspace(globalfeetypes.ModuleName)
		if !ok {
			panic("global fee params subspace not found")
		}
		globlaFeeParamsSubspace.SetParamSet(ctx, &globalFeeParams)
		// -- --

		// -- tariff params --
		tariffParamsSubspace, ok := paramauthoritykeeper.GetSubspace(tarifftypes.ModuleName)
		if !ok {
			panic("tariff params subspace not found")
		}
		paramAuth := paramauthoritykeeper.GetAuthority(ctx)
		distributionEntities := []tarifftypes.DistributionEntity{
			{
				Address: paramAuth,
				Share:   sdk.NewDec(1),
			},
		}
		tariffParams := tarifftypes.Params{
			Share:                sdk.NewDecWithPrec(8, 1),
			DistributionEntities: distributionEntities,
			TransferFeeBps:       sdk.OneInt(),
			TransferFeeMax:       sdk.NewInt(5000000),
			TransferFeeDenom:     feeDenom.Denom,
		}
		tariffParamsSubspace.SetParamSet(ctx, &tariffParams)
		// -- --

		return versionMap, err
	}
}
