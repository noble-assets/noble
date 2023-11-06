package krypton

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	upgradeTypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	connectionkeeper "github.com/cosmos/ibc-go/v4/modules/core/03-connection/keeper"
	connectiontypes "github.com/cosmos/ibc-go/v4/modules/core/03-connection/types"
	consumerKeeper "github.com/cosmos/interchain-security/v2/x/ccv/consumer/keeper"
	consumertypes "github.com/cosmos/interchain-security/v2/x/ccv/consumer/types"
	"os"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.Codec,
	consumerKeeper consumerKeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
) upgradeTypes.UpgradeHandler {
	// The below is taken from https://github.com/cosmos/interchain-security/blob/v2.0.0/app/consumer-democracy/app.go#L635-L672.
	return func(ctx sdk.Context, _ upgradeTypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		connectionKeeper.SetParams(ctx, connectiontypes.DefaultParams())

		fromVM := make(map[string]uint64)

		for moduleName, eachModule := range mm.Modules {
			fromVM[moduleName] = eachModule.ConsensusVersion()
		}

		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			return fromVM, err
		}
		nodeHome := userHomeDir + "/.sovereign/config/genesis.json"
		appState, _, err := genutiltypes.GenesisStateFromGenFile(nodeHome)
		if err != nil {
			return fromVM, fmt.Errorf("failed to unmarshal genesis state: %w", err)
		}

		consumerGenesis := consumertypes.GenesisState{}
		cdc.MustUnmarshalJSON(appState[consumertypes.ModuleName], &consumerGenesis)

		consumerGenesis.PreCCV = true
		consumerKeeper.InitGenesis(ctx, &consumerGenesis)

		ctx.Logger().Info("start to run module migrations...")

		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
