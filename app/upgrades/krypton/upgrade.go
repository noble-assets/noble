package krypton

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	consumerKeeper "github.com/cosmos/interchain-security/v2/x/ccv/consumer/keeper"
	consumerTypes "github.com/cosmos/interchain-security/v2/x/ccv/consumer/types"
	"github.com/spf13/cast"
	fiatTokenFactoryKeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	options servertypes.AppOptions,
	consumerKeeper consumerKeeper.Keeper,
	fiatTokenFactoryKeeper *fiatTokenFactoryKeeper.Keeper,
) upgradeTypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradeTypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		home := cast.ToString(options.Get(flags.FlagHome))
		state, _, err := genutiltypes.GenesisStateFromGenFile(home + "/config/genesis.json")
		if err != nil {
			return vm, err
		}

		var genesis consumerTypes.GenesisState
		cdc.MustUnmarshalJSON(state[consumerTypes.ModuleName], &genesis)

		genesis.PreCCV = true
		genesis.Params.SoftOptOutThreshold = "0.05"
		genesis.Params.RewardDenoms = []string{
			fiatTokenFactoryKeeper.GetMintingDenom(ctx).Denom, // USDC
		}

		consumerKeeper.InitGenesis(ctx, &genesis)

		switch ctx.ChainID() {
		case TestnetChainID:
			// TODO
		case MainnetChainID:
			consumerKeeper.SetDistributionTransmissionChannel(ctx, "channel-4")
		}

		return vm, nil
	}
}
