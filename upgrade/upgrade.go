// SPDX-License-Identifier: Apache-2.0
//
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
	"encoding/json"
	"errors"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dollarkeeper "dollar.noble.xyz/v2/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	orbitertypes "github.com/noble-assets/orbiter/types"
	orbitercore "github.com/noble-assets/orbiter/types/core"
)

// orbiterCustomGen overriede the default module genesis to pause the Hyperlane forwarding.
func orbiterCustomGen(cdc codec.Codec) json.RawMessage {
	gen := orbitertypes.DefaultGenesisState()
	gen.ForwarderGenesis.PausedProtocolIds = append(gen.ForwarderGenesis.PausedProtocolIds, orbitercore.PROTOCOL_HYPERLANE)

	return cdc.MustMarshalJSON(gen)
}

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	cdc codec.Codec,
	logger log.Logger,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		if module, ok := mm.Modules[orbitercore.ModuleName].(module.HasGenesis); ok {
			module.InitGenesis(sdkCtx, cdc, orbiterCustomGen(cdc))
		}

		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		err = ClaimDistributionFunds(ctx, logger, addressCodec, authorityKeeper, bankKeeper)
		if err != nil {
			return vm, err
		}

		err = UpdateVaultsState(ctx, addressCodec, authorityKeeper, dollarKeeper)
		if err != nil {
			return vm, err
		}

		// The IBC light client for the shido_9008-1 chain has expired on
		// Noble's mainnet. In IBC-Go v8.7.0, the MsgRecoverClient message does
		// not support the LegacyAminoJSON signing mode, preventing recovery
		// via the Noble Maintenance Multisig. As a result, the client must be
		// manually recovered as part of this software upgrade.
		err = clientKeeper.RecoverClient(sdkCtx, "07-tendermint-106", "07-tendermint-186")
		if err != nil {
			logger.Error("unable to recover shido_9008-1 light client", "err", err)
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

// ClaimDistributionFunds transfers all transaction fees accrued by Noble prior
// to the v8 Helium upgrade (November 2024) to the x/authority owner. The funds
// are currently stuck as the x/distribution module was removed and replaced by
// the x/authority module without a proper migration of funds.
func ClaimDistributionFunds(ctx context.Context, logger log.Logger, addressCodec address.Codec, authorityKeeper *authoritykeeper.Keeper, bankKeeper bankkeeper.Keeper) error {
	// NOTE: We hardcode the x/distribution module name to avoid an import.
	address := authtypes.NewModuleAddress("distribution")
	balance := bankKeeper.GetAllBalances(ctx, address)
	if balance.IsZero() {
		// We return early in the case that there are no claimable funds.
		return nil
	}

	authority, err := authorityKeeper.Owner.Get(ctx)
	if err != nil {
		return errors.New("unable to get underlying authority address from state")
	}
	authorityBz, err := addressCodec.StringToBytes(authority)
	if err != nil {
		return errors.New("unable to decode underlying authority address")
	}

	err = bankKeeper.SendCoins(ctx, address, authorityBz, balance)
	if err != nil {
		return errors.New("unable to transfer stuck distribution funds")
	}

	logger.Info("claimed stuck distribution module funds", "amount", balance.String())

	return nil
}

// UpdateVaultsState sets state variables around Vaults Season One and Season
// Two. We do this so that we can remove these values from the app.yaml file,
// allowing us to ship one binary for both mainnet and testnet.
func UpdateVaultsState(ctx context.Context, addressCodec address.Codec, authorityKeeper *authoritykeeper.Keeper, dollarKeeper *dollarkeeper.Keeper) error {
	switch sdk.UnwrapSDKContext(ctx).ChainID() {
	case TestnetChainID:
		err := dollarKeeper.VaultsSeasonOneEnded.Set(ctx, true)
		if err != nil {
			return errors.New("unable to mark vaults season one as ended")
		}

		authority, err := authorityKeeper.Owner.Get(ctx)
		if err != nil {
			return errors.New("unable to get underlying authority address from state")
		}
		authorityBz, err := addressCodec.StringToBytes(authority)
		if err != nil {
			return errors.New("unable to decode underlying authority address")
		}
		err = dollarKeeper.VaultsSeasonTwoYieldCollector.Set(ctx, authorityBz)
		if err != nil {
			return errors.New("unable to set vaults season two yield collector")
		}
	case MainnetChainID:
		// NOTE: Vaults Season One has already been marked as ended on mainnet
		// via the v10.1 Ember upgrade, so we safely skip that update here.

		yieldCollector, err := addressCodec.StringToBytes("noble17m7dleu26hgwk842hrvfmh8mvrtp7p68k4zq8l")
		if err != nil {
			return errors.New("unable to decode vaults season two yield collector")
		}
		err = dollarKeeper.VaultsSeasonTwoYieldCollector.Set(ctx, yieldCollector)
		if err != nil {
			return errors.New("unable to set vaults season two yield collector")
		}
	}

	return nil
}
