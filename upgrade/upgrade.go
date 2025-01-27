// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upgrade

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	capabilityKeeper *capabilitykeeper.Keeper,
	consensusKeeper consensuskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return nil, err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		FixICS27ChannelCapabilities(sdkCtx, capabilityKeeper)

		err = EnableVoteExtensions(ctx, logger, plan, consensusKeeper)
		if err != nil {
			return nil, err
		}

		logger.Info("Welcome to a new generation of Noble!" + UpgradeASCII)

		return vm, nil
	}
}

// FixICS27ChannelCapabilities finds all capabilities wrongfully owned by the
// ICA Controller module and replaces them with the ICA Host module. This was
// introduced in the v8 Helium upgrade after we executed the recommended ICS27
// migration logic for chains that utilize the ICA Controller module.
func FixICS27ChannelCapabilities(ctx sdk.Context, capabilityKeeper *capabilitykeeper.Keeper) {
	index := capabilityKeeper.GetLatestIndex(ctx)

	for i := uint64(1); i < index; i++ {
		wrapper, ok := capabilityKeeper.GetOwners(ctx, i)
		if !ok {
			continue
		}

		for _, owner := range wrapper.GetOwners() {
			if owner.Module == icacontrollertypes.SubModuleName {
				wrapper.Remove(owner)
				wrapper.Set(capabilitytypes.Owner{
					Module: icahosttypes.SubModuleName,
					Name:   owner.Name,
				})
			}
		}

		capabilityKeeper.SetOwners(ctx, i, wrapper)
	}

	capabilityKeeper.InitMemStore(ctx)
}

// EnableVoteExtensions updates the consensus parameters of Noble to enable
// vote extensions one block after the Argentum upgrade is performed.
func EnableVoteExtensions(ctx context.Context, logger log.Logger, plan upgradetypes.Plan, consensusKeeper consensuskeeper.Keeper) error {
	var params cmtproto.ConsensusParams
	var err error

	params, err = consensusKeeper.ParamsStore.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get consensus params from state")
	}

	params.Abci = &tmtypes.ABCIParams{
		VoteExtensionsEnableHeight: plan.Height + 1,
	}

	err = consensusKeeper.ParamsStore.Set(ctx, params)
	if err != nil {
		return errors.Wrap(err, "unable to set consensus params to state")
	}

	logger.Info(fmt.Sprintf("enabling vote extensions at block %d", params.Abci.VoteExtensionsEnableHeight))

	return nil
}
