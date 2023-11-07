package krypton

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	consumerkeeper "github.com/cosmos/interchain-security/v2/x/ccv/consumer/keeper"
	consumertypes "github.com/cosmos/interchain-security/v2/x/ccv/consumer/types"
	fiattokenfactorykeeper "github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	options servertypes.AppOptions,
	consumerKeeper consumerkeeper.Keeper,
	fiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		bz, err := os.ReadFile(fmt.Sprintf("%s/config/ccv.json", options.Get(flags.FlagHome)))
		if err != nil {
			return vm, err
		}

		var genesis consumertypes.GenesisState
		cdc.MustUnmarshalJSON(bz, &genesis)

		genesis.PreCCV = true
		genesis.Params.SoftOptOutThreshold = "0.05"
		genesis.Params.RewardDenoms = []string{
			fiatTokenFactoryKeeper.GetMintingDenom(ctx).Denom, // USDC
		}

		consumerKeeper.InitGenesis(ctx, &genesis)

		// TODO: Should we set the distribution channel?

		return vm, nil
	}
}
