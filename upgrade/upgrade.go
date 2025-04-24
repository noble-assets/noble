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
	"errors"
	"fmt"
	"sort"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	dollarkeeper "dollar.noble.xyz/keeper"
	dollartypes "dollar.noble.xyz/types"
	ismkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdhkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/keeper"
	pdhtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	hyperlanekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	hyperlanetypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	authoritykeeper "github.com/noble-assets/authority/keeper"
	authoritytypes "github.com/noble-assets/authority/types"
	swapkeeper "swap.noble.xyz/keeper"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	logger log.Logger,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	hyperlaneKeeper *hyperlanekeeper.Keeper,
	swapKeeper *swapkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		vm, err := mm.RunMigrations(ctx, cfg, vm)
		if err != nil {
			return vm, err
		}

		err = ClaimSwapPoolsYield(ctx, logger, addressCodec, authorityKeeper, bankKeeper, dollarKeeper, swapKeeper)
		if err != nil {
			return vm, err
		}

		err = InitializeHyperlaneModule(ctx, logger, addressCodec, hyperlaneKeeper)
		if err != nil {
			return vm, err
		}

		logger.Info(UpgradeASCII)

		return vm, nil
	}
}

// ClaimSwapPoolsYield claims the $USDN yield accrued inside the Noble Swap
// pools and sends it to the authority address.
func ClaimSwapPoolsYield(
	ctx context.Context,
	logger log.Logger,
	addressCodec address.Codec,
	authorityKeeper *authoritykeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	dollarKeeper *dollarkeeper.Keeper,
	swapKeeper *swapkeeper.Keeper,
) error {
	authority, err := authorityKeeper.Owner.Get(ctx)
	if err != nil {
		return errors.New("unable to get underlying authority address from state")
	}
	authorityBz, err := addressCodec.StringToBytes(authority)
	if err != nil {
		return errors.New("unable to decode underlying authority address")
	}

	dollarServer := dollarkeeper.NewMsgServer(dollarKeeper)

	pools := swapKeeper.GetPools(ctx)
	for _, pool := range pools {
		yield, address, err := dollarKeeper.GetYield(ctx, pool.Address)
		if err != nil {
			return fmt.Errorf("unable to get yield for pool %d", pool.Id)
		}

		_, err = dollarServer.ClaimYield(ctx, &dollartypes.MsgClaimYield{Signer: pool.Address})
		if err != nil {
			return fmt.Errorf("unable to claim yield for pool %d", pool.Id)
		}

		err = bankKeeper.SendCoins(ctx, address, authorityBz, sdk.NewCoins(sdk.NewCoin(dollarKeeper.GetDenom(), yield)))
		if err != nil {
			return fmt.Errorf("unable to transfer yield for pool %d", pool.Id)
		}

		logger.Info("claimed swap pool yield", "pool", pool.Id, "yield", yield)
	}

	return nil
}

// InitializeHyperlaneModule creates a default Hyperlane ISM and Mailbox.
func InitializeHyperlaneModule(
	ctx context.Context,
	logger log.Logger,
	addressCodec address.Codec,
	hyperlaneKeeper *hyperlanekeeper.Keeper,
) error {
	chainId := sdk.UnwrapSDKContext(ctx).ChainID()

	var localDomain uint32
	switch chainId {
	case TestnetChainID:
		localDomain = TestnetHyperlaneDomain
	case MainnetChainID:
		localDomain = MainnetHyperlaneDomain
	default:
		return fmt.Errorf("cannot initialize hyperlane module on %s chain", chainId)
	}

	authority, err := addressCodec.BytesToString(authoritytypes.ModuleAddress)
	if err != nil {
		return errors.New("unable to encode authority address")
	}

	ismServer := ismkeeper.NewMsgServerImpl(&hyperlaneKeeper.IsmKeeper)
	pdhServer := pdhkeeper.NewMsgServerImpl(&hyperlaneKeeper.PostDispatchKeeper)
	hyperlaneServer := hyperlanekeeper.NewMsgServerImpl(hyperlaneKeeper)

	createRoutingIsmRes, err := ismServer.CreateRoutingIsm(ctx, &ismtypes.MsgCreateRoutingIsm{Creator: authority})
	if err != nil {
		return fmt.Errorf("unable to create routing ism: %w", err)
	}
	ismId := createRoutingIsmRes.Id
	logger.Info("created noble hyperlane ism", "id", ismId)

	if isms, found := HyperlaneDefaultISMs[chainId]; found {
		for _, ism := range isms {
			sort.Strings(ism.Validators)

			res, err := ismServer.CreateMerkleRootMultisigIsm(ctx, &ismtypes.MsgCreateMerkleRootMultisigIsm{
				Creator:    authority,
				Validators: ism.Validators,
				Threshold:  ism.Threshold,
			})
			if err != nil {
				return fmt.Errorf("unable to create default ism for domain %d: %w", ism.Domain, err)
			}
			underlyingIsmId := res.Id
			logger.Info(fmt.Sprintf("created default hyperlane ism for %s", ism.Name), "domain", ism.Domain, "id", underlyingIsmId)

			_, err = ismServer.SetRoutingIsmDomain(ctx, &ismtypes.MsgSetRoutingIsmDomain{
				IsmId: ismId,
				Route: ismtypes.Route{
					Ism:    underlyingIsmId,
					Domain: ism.Domain,
				},
				Owner: authority,
			})
			if err != nil {
				return fmt.Errorf("unable to set default ism in routing ism: %w", err)
			}
		}
	}

	createMailboxRes, err := hyperlaneServer.CreateMailbox(ctx, &hyperlanetypes.MsgCreateMailbox{
		Owner:       authority,
		LocalDomain: localDomain,
		DefaultIsm:  ismId,
	})
	if err != nil {
		return fmt.Errorf("unable to create mailbox: %w", err)
	}
	mailboxId := createMailboxRes.Id
	logger.Info("created noble hyperlane mailbox", "id", mailboxId)

	createMerkleTreeHookRes, err := pdhServer.CreateMerkleTreeHook(ctx, &pdhtypes.MsgCreateMerkleTreeHook{
		Owner:     authority,
		MailboxId: mailboxId,
	})
	if err != nil {
		return fmt.Errorf("unable to create merkle tree hook: %w", err)
	}
	requiredHook := createMerkleTreeHookRes.Id
	logger.Info("created noble hyperlane merkle tree hook", "id", requiredHook)

	_, err = hyperlaneServer.SetMailbox(ctx, &hyperlanetypes.MsgSetMailbox{
		Owner:        authority,
		MailboxId:    mailboxId,
		RequiredHook: &requiredHook,
	})
	if err != nil {
		return fmt.Errorf("unable to set mailbox: %w", err)
	}

	return nil
}
