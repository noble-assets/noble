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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	controllertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	hosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	"github.com/docker/docker/client"
	florintypes "github.com/monerium/module-noble/v2/types"
	authoritytypes "github.com/noble-assets/authority/types"
	halotypes "github.com/noble-assets/halo/v2/types"
	"github.com/noble-assets/noble/upgrade"
	auratypes "github.com/ondoprotocol/usdy-noble/v2/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	ghcrRepo        = "ghcr.io/noble-assets/noble"
	containerUidGid = "1025:1025"
	e2eChainID      = upgrade.MainnetChainID
)

var (
	LocalImages = []ibc.DockerImage{
		{
			Repository: "noble",
			Version:    "local",
			UIDGID:     "1025:1025",
		},
	}

	DenomMetadataUsdc = banktypes.Metadata{
		Description: "USD Coin",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "uusdc",
				Exponent: 0,
				Aliases: []string{
					"microusdc",
				},
			},
			{
				Denom:    "usdc",
				Exponent: 6,
				Aliases:  []string{},
			},
		},
		Base:    "uusdc",
		Display: "usdc",
		Name:    "usdc",
		Symbol:  "USDC",
	}
)

type NobleWrapper struct {
	Chain       *cosmos.CosmosChain
	FiatTfRoles FiatTfRoles
	CCTPRoles   CCTPRoles
	Authority   ibc.Wallet
}

type FiatTfRoles struct {
	Owner            ibc.Wallet
	MasterMinter     ibc.Wallet
	MinterController ibc.Wallet
	Minter           ibc.Wallet
	Blacklister      ibc.Wallet
	Pauser           ibc.Wallet
}

type CCTPRoles struct {
	Owner           ibc.Wallet
	AttesterManager ibc.Wallet
	TokenController ibc.Wallet
}

func NobleEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	icatypes.RegisterInterfaces(cfg.InterfaceRegistry)
	hosttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	controllertypes.RegisterInterfaces(cfg.InterfaceRegistry)

	// register custom types
	fiattokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	cctptypes.RegisterInterfaces(cfg.InterfaceRegistry)
	halotypes.RegisterInterfaces(cfg.InterfaceRegistry)
	auratypes.RegisterInterfaces(cfg.InterfaceRegistry)
	florintypes.RegisterInterfaces(cfg.InterfaceRegistry)
	authoritytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

func SimdEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	icatypes.RegisterInterfaces(cfg.InterfaceRegistry)
	hosttypes.RegisterInterfaces(cfg.InterfaceRegistry)
	controllertypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func NobleChainSpec(
	ctx context.Context,
	nw *NobleWrapper,
	chainID string,
	version []ibc.DockerImage,
	nv, nf int,
	setupAllCircleRoles bool,
) *interchaintest.ChainSpec {
	return &interchaintest.ChainSpec{
		NumValidators: &nv,
		NumFullNodes:  &nf,
		ChainConfig: ibc.ChainConfig{
			Type:           "cosmos",
			Name:           "noble",
			ChainID:        chainID,
			Bin:            "nobled",
			Denom:          "ustake",
			Bech32Prefix:   "noble",
			CoinType:       "118",
			GasPrices:      "0.0ustake",
			GasAdjustment:  1.1,
			TrustingPeriod: "504h",
			NoHostMount:    false,
			Images:         version,
			EncodingConfig: NobleEncoding(),
			PreGenesis:     preGenesisAll(ctx, nw, setupAllCircleRoles),
			ModifyGenesis:  modifyGenesisAll(nw, setupAllCircleRoles),
		},
	}
}

// modifyGenesisAll modifies the genesis file to with fields needed to start chain
//
// setupAllCircleRoles: if true, all Tokenfactory and CCTP roles will be setup and tied to a wallet at genesis,
// if false, only the Owner role will be setup
func modifyGenesisAll(nw *NobleWrapper, setupAllCircleRoles bool) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		updatedGenesis := []cosmos.GenesisKV{
			cosmos.NewGenesisKV("app_state.bank.denom_metadata", []banktypes.Metadata{DenomMetadataUsdc}),
			cosmos.NewGenesisKV("app_state.fiat-tokenfactory.owner", fiattokenfactorytypes.Owner{Address: nw.FiatTfRoles.Owner.FormattedAddress()}),
			cosmos.NewGenesisKV("app_state.fiat-tokenfactory.paused", fiattokenfactorytypes.Paused{Paused: false}),
			cosmos.NewGenesisKV("app_state.fiat-tokenfactory.mintingDenom", fiattokenfactorytypes.MintingDenom{Denom: DenomMetadataUsdc.Base}),
			cosmos.NewGenesisKV("app_state.staking.params.bond_denom", "ustake"),
			cosmos.NewGenesisKV("app_state.interchainaccounts.host_genesis_state.params.allow_messages", []string{"*"}),
		}

		// Modify the genesis file with the authority address for the appropriate module.
		// Prior to v8.0.0 the SL paramauthoritymodule was used, and post v8.0.0 the Noble authority module is used.
		if cc.Images[0].Version == "v7.0.0" {
			updatedGenesis = append(updatedGenesis, cosmos.NewGenesisKV("app_state.params.params.authority", nw.Authority.FormattedAddress()))
			updatedGenesis = append(updatedGenesis, cosmos.NewGenesisKV("app_state.upgrade.params.authority", nw.Authority.FormattedAddress()))
		} else {
			updatedGenesis = append(updatedGenesis, cosmos.NewGenesisKV("app_state.authority.owner", nw.Authority.FormattedAddress()))
		}

		// Modify the genesis file with the appropriate Dollar Vaults Season One and Two state.
		// For v10.1, we opt to not set this state as these values were hardcoded in the app wiring!
		if cc.Images[0].Version != "v10.1.2" {
			updatedGenesis = append(updatedGenesis, cosmos.NewGenesisKV("app_state.dollar.vaults.season_one_ended", true))
			updatedGenesis = append(updatedGenesis, cosmos.NewGenesisKV("app_state.dollar.vaults.season_two_yield_collector", nw.Authority.FormattedAddress()))
		}

		if setupAllCircleRoles {
			allFiatTFRoles := []cosmos.GenesisKV{
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.masterMinter", fiattokenfactorytypes.MasterMinter{Address: nw.FiatTfRoles.MasterMinter.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.mintersList", []fiattokenfactorytypes.Minters{{Address: nw.FiatTfRoles.Minter.FormattedAddress(), Allowance: sdk.Coin{Denom: DenomMetadataUsdc.Base, Amount: math.NewInt(100_00_000)}}}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.pauser", fiattokenfactorytypes.Pauser{Address: nw.FiatTfRoles.Pauser.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.blacklister", fiattokenfactorytypes.Blacklister{Address: nw.FiatTfRoles.Blacklister.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.masterMinter", fiattokenfactorytypes.MasterMinter{Address: nw.FiatTfRoles.MasterMinter.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.minterControllerList", []fiattokenfactorytypes.MinterController{{Minter: nw.FiatTfRoles.Minter.FormattedAddress(), Controller: nw.FiatTfRoles.MinterController.FormattedAddress()}}),
				cosmos.NewGenesisKV("app_state.cctp", cctptypes.GenesisState{
					Owner:                             nw.CCTPRoles.Owner.FormattedAddress(),
					AttesterManager:                   nw.CCTPRoles.AttesterManager.FormattedAddress(),
					TokenController:                   nw.CCTPRoles.TokenController.FormattedAddress(),
					BurningAndMintingPaused:           &cctptypes.BurningAndMintingPaused{Paused: false},
					SendingAndReceivingMessagesPaused: &cctptypes.SendingAndReceivingMessagesPaused{Paused: false},
					NextAvailableNonce:                &cctptypes.Nonce{Nonce: 0},
					SignatureThreshold:                &cctptypes.SignatureThreshold{Amount: 2},
				}),
			}
			updatedGenesis = append(updatedGenesis, allFiatTFRoles...)
		}

		return cosmos.ModifyGenesis(updatedGenesis)(cc, b)
	}
}

func preGenesisAll(ctx context.Context, nw *NobleWrapper, setupAllCircleRoles bool) func(ibc.Chain) error {
	return func(cc ibc.Chain) (err error) {
		val := nw.Chain.Validators[0]

		nw.FiatTfRoles, err = createTokenfactoryRoles(ctx, val, setupAllCircleRoles)
		if err != nil {
			return err
		}

		nw.Authority, err = createAuthorityRole(ctx, val)
		if err != nil {
			return err
		}

		nw.CCTPRoles, err = createCCTPRoles(ctx, val)
		if err != nil {
			return err
		}

		return err
	}
}

// createTokenfactoryRoles Creates wallets to be tied to TF roles with 0 amount. Meant to run pre-genesis.
// After creating thw wallets, it recovers the key on the specified validator.
//
// setupAllCircleRoles: if true, a wallet for all Tokenfactory and CCTP roles will be created,
// if false, only the Owner role will be created
func createTokenfactoryRoles(ctx context.Context, val *cosmos.ChainNode, setupAllCircleRoles bool) (FiatTfRoles, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error
	fiatTfRoles := FiatTfRoles{}

	fiatTfRoles.Owner, err = nobleVal.BuildRelayerWallet(ctx, "owner-fiatTF")
	if err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	if err := val.RecoverKey(ctx, fiatTfRoles.Owner.KeyName(), fiatTfRoles.Owner.Mnemonic()); err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to restore %s wallet: %w", fiatTfRoles.Owner.KeyName(), err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: fiatTfRoles.Owner.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  math.ZeroInt(),
	}
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []sdk.Coin{sdk.NewCoin(genesisWallet.Denom, genesisWallet.Amount)})
	if err != nil {
		return FiatTfRoles{}, err
	}

	if !setupAllCircleRoles {
		return fiatTfRoles, nil
	}

	fiatTfRoles.MasterMinter, err = nobleVal.BuildRelayerWallet(ctx, "masterminter")
	if err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to create %s wallet: %w", "masterminter", err)
	}
	fiatTfRoles.MinterController, err = nobleVal.BuildRelayerWallet(ctx, "mintercontroller")
	if err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to create %s wallet: %w", "mintercontroller", err)
	}
	fiatTfRoles.Minter, err = nobleVal.BuildRelayerWallet(ctx, "minter")
	if err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to create %s wallet: %w", "minter", err)
	}
	fiatTfRoles.Blacklister, err = nobleVal.BuildRelayerWallet(ctx, "blacklister")
	if err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to create %s wallet: %w", "blacklister", err)
	}
	fiatTfRoles.Pauser, err = nobleVal.BuildRelayerWallet(ctx, "pauser")
	if err != nil {
		return FiatTfRoles{}, fmt.Errorf("failed to create %s wallet: %w", "pauser", err)
	}

	walletsToRestore := []ibc.Wallet{fiatTfRoles.MasterMinter, fiatTfRoles.MinterController, fiatTfRoles.Minter, fiatTfRoles.Blacklister, fiatTfRoles.Pauser}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return FiatTfRoles{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: fiatTfRoles.MasterMinter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: fiatTfRoles.MinterController.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: fiatTfRoles.Minter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: fiatTfRoles.Blacklister.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: fiatTfRoles.Pauser.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []sdk.Coin{sdk.NewCoin(wallet.Denom, wallet.Amount)})
		if err != nil {
			return FiatTfRoles{}, err
		}
	}

	return fiatTfRoles, nil
}

func createAuthorityRole(ctx context.Context, val *cosmos.ChainNode) (ibc.Wallet, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	authority, err := nobleVal.BuildRelayerWallet(ctx, "authority")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	if err := val.RecoverKey(ctx, authority.KeyName(), authority.Mnemonic()); err != nil {
		return nil, fmt.Errorf("failed to restore %s wallet: %w", authority.KeyName(), err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: authority.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  math.ZeroInt(),
	}
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []sdk.Coin{sdk.NewCoin(genesisWallet.Denom, genesisWallet.Amount)})
	if err != nil {
		return nil, err
	}

	return authority, nil
}

func createCCTPRoles(ctx context.Context, val *cosmos.ChainNode) (CCTPRoles, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	cctpRoles := CCTPRoles{}

	cctpRoles.Owner, err = nobleVal.BuildRelayerWallet(ctx, "cctp-owner")
	if err != nil {
		return CCTPRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	cctpRoles.AttesterManager, err = nobleVal.BuildRelayerWallet(ctx, "attester-manager")
	if err != nil {
		return CCTPRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	cctpRoles.TokenController, err = nobleVal.BuildRelayerWallet(ctx, "token-controller")
	if err != nil {
		return CCTPRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	walletsToRestore := []ibc.Wallet{cctpRoles.Owner, cctpRoles.AttesterManager, cctpRoles.TokenController}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return CCTPRoles{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: cctpRoles.Owner.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: cctpRoles.AttesterManager.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: cctpRoles.TokenController.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []sdk.Coin{sdk.NewCoin(wallet.Denom, wallet.Amount)})
		if err != nil {
			return CCTPRoles{}, err
		}
	}

	return cctpRoles, nil
}

// NobleSpinUp starts noble chain
//
// setupAllCircleRoles: if true, all Tokenfactory and CCTP roles will be created and setup at genesis,
// if false, only the Owner role will be created
func NobleSpinUp(t *testing.T, ctx context.Context, version []ibc.DockerImage, setupAllCircleRoles bool) (nw NobleWrapper, client *client.Client) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("noble", "noblepub")

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	numValidators := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		NobleChainSpec(ctx, &nw, e2eChainID, version, numValidators, numFullNodes, setupAllCircleRoles),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	nw.Chain = chains[0].(*cosmos.CosmosChain)
	noble := nw.Chain

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	return
}

// NobleSpinUpIBC is the same as nobleSpinUp but it also spins up a ibcSimd chain and creates
// an IBC path between them
//
// setupAllCircleRoles: if true, all Tokenfactory and CCTP roles will be created and setup at genesis,
// if false, only the Owner role will be created
func NobleSpinUpIBC(t *testing.T, ctx context.Context, version []ibc.DockerImage, setupAllCircleRoles bool) (
	nw NobleWrapper,
	ibcSimd *cosmos.CosmosChain,
	rf interchaintest.RelayerFactory,
	r ibc.Relayer,
	ibcPathName string,
	rep *testreporter.Reporter,
	eRep *testreporter.RelayerExecReporter,
	client *client.Client,
	network string,
) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("noble", "noblepub")

	rep = testreporter.NewNopReporter()
	eRep = rep.RelayerExecReporter(t)

	client, network = interchaintest.DockerSetup(t)

	numValidators := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		NobleChainSpec(ctx, &nw, e2eChainID, version, numValidators, numFullNodes, setupAllCircleRoles),
		{
			Name:    "ibc-go-simd",
			Version: "v8.7.0",
			ChainConfig: ibc.ChainConfig{
				EncodingConfig: SimdEncoding(),
			},
			NumValidators: &numValidators,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	nw.Chain = chains[0].(*cosmos.CosmosChain)
	noble := nw.Chain
	ibcSimd = chains[1].(*cosmos.CosmosChain)

	rf = interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t))
	r = rf.Build(t, client, network)

	ibcPathName = "path"
	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(ibcSimd).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  ibcSimd,
			Relayer: r,
			Path:    ibcPathName,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	return
}

////////////////////////////////
// Fiat Token Factory Helpers //
////////////////////////////////

// BlacklistAccount blacklists an account and then runs the `show-blacklisted` query to ensure the
// account was successfully blacklisted on chain
func BlacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, toBlacklist ibc.Wallet) {
	_, err := val.ExecTx(ctx, blacklister.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", toBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to query show-blacklisted")

	var showBlacklistedResponse fiattokenfactorytypes.QueryGetBlacklistedResponse
	err = json.Unmarshal(res, &showBlacklistedResponse)
	require.NoError(t, err, "failed to unmarshal show-blacklisted response")

	expectedBlacklistResponse := fiattokenfactorytypes.QueryGetBlacklistedResponse{
		Blacklisted: fiattokenfactorytypes.Blacklisted{
			AddressBz: toBlacklist.Address(),
		},
	}

	require.Equal(t, expectedBlacklistResponse.Blacklisted, showBlacklistedResponse.Blacklisted)
}

// UnblacklistAccount unblacklists an account and then runs the `show-blacklisted` query to ensure the
// account was successfully unblacklisted on chain
func UnblacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, unBlacklist ibc.Wallet) {
	_, err := val.ExecTx(ctx, blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", unBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", unBlacklist.FormattedAddress())
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
}

// PauseFiatTF pauses the fiat tokenfactory. It then runs the `show-paused` query to ensure the
// the tokenfactory was successfully paused
func PauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser ibc.Wallet) {
	_, err := val.ExecTx(ctx, pauser.KeyName(), "fiat-tokenfactory", "pause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")

	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err)

	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: true,
		},
	}
	require.Equal(t, expectedPaused, showPausedResponse)
}

// UnpauseFiatTF pauses the fiat tokenfactory. It then runs the `show-paused` query to ensure the
// the tokenfactory was successfully unpaused
func UnpauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser ibc.Wallet) {
	_, err := val.ExecTx(ctx, pauser.KeyName(), "fiat-tokenfactory", "unpause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")

	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err, "failed to unmarshal show-paused response")

	expectedUnpaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: false,
		},
	}
	require.Equal(t, expectedUnpaused, showPausedResponse)
}

// SetupMinterAndController creates a minter controller and minter. It also sets up a minter with an specified allowance of `uusdc`
func SetupMinterAndController(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, val *cosmos.ChainNode, masterMinter ibc.Wallet, allowance int64) (minter ibc.Wallet, minterController ibc.Wallet) {
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	minterController = w[0]
	minter = w[1]

	_, err := val.ExecTx(ctx, masterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController.FormattedAddress(), minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMC, err := ShowMinterController(ctx, val, minterController)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter.FormattedAddress(),
			Controller: minterController.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMC.MinterController)

	ConfigureMinter(t, ctx, val, minterController, minter, allowance)

	return minter, minterController
}

// ConfigureMinter configures a minter with a specified allowance of `uusdc`. It then runs the `show-minters` query to ensure
// the minter was properly configured
func ConfigureMinter(t *testing.T, ctx context.Context, val *cosmos.ChainNode, minterController, minter ibc.Wallet, allowance int64) {
	_, err := val.ExecTx(ctx, minterController.KeyName(), "fiat-tokenfactory", "configure-minter", minter.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.NoError(t, err, "error configuring minter")

	showMinter, err := ShowMinters(ctx, val, minter)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter.FormattedAddress(),
			Allowance: sdk.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinter.Minters)
}

// ShowMinterController queries for a specific minter controller by running: `query fiat-tokenfactory show-minter-controller <address>`.
// An error is returned if the minter controller does not exist
func ShowMinterController(ctx context.Context, val *cosmos.ChainNode, minterController ibc.Wallet) (fiattokenfactorytypes.QueryGetMinterControllerResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController.FormattedAddress())
	if err != nil {
		return fiattokenfactorytypes.QueryGetMinterControllerResponse{}, err
	}

	var showMinterController fiattokenfactorytypes.QueryGetMinterControllerResponse
	err = json.Unmarshal(res, &showMinterController)
	if err != nil {
		return fiattokenfactorytypes.QueryGetMinterControllerResponse{}, err
	}

	return showMinterController, nil
}

// ShowMinters queries for a specific minter by running: `query fiat-tokenfactory show-minters <address>`.
// An error is returned if the minter does not exist
func ShowMinters(ctx context.Context, val *cosmos.ChainNode, minter ibc.Wallet) (fiattokenfactorytypes.QueryGetMintersResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter.FormattedAddress())
	if err != nil {
		return fiattokenfactorytypes.QueryGetMintersResponse{}, err
	}

	var showMinters fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &showMinters)
	if err != nil {
		return fiattokenfactorytypes.QueryGetMintersResponse{}, err
	}

	return showMinters, nil
}

// ShowOwner queries for the token factory Owner by running: `query fiat-tokenfactory show-owner`.
func ShowOwner(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetOwnerResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-owner")
	if err != nil {
		return fiattokenfactorytypes.QueryGetOwnerResponse{}, err
	}

	var showOwnerResponse fiattokenfactorytypes.QueryGetOwnerResponse
	err = json.Unmarshal(res, &showOwnerResponse)
	if err != nil {
		return fiattokenfactorytypes.QueryGetOwnerResponse{}, err
	}

	return showOwnerResponse, nil
}

// ShowMasterMinter queries for the token factory Master Minter by running: `query fiat-tokenfactory show-master-minter`.
func ShowMasterMinter(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetMasterMinterResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-master-minter")
	if err != nil {
		return fiattokenfactorytypes.QueryGetMasterMinterResponse{}, err
	}

	var showMMResponse fiattokenfactorytypes.QueryGetMasterMinterResponse
	err = json.Unmarshal(res, &showMMResponse)
	if err != nil {
		return fiattokenfactorytypes.QueryGetMasterMinterResponse{}, err
	}

	return showMMResponse, nil
}

// ShowPauser queries for the token factory Pauser by running: `query fiat-tokenfactory show-pauser`.
func ShowPauser(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetPauserResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-pauser")
	if err != nil {
		return fiattokenfactorytypes.QueryGetPauserResponse{}, err
	}

	var showPauserRes fiattokenfactorytypes.QueryGetPauserResponse
	err = json.Unmarshal(res, &showPauserRes)
	if err != nil {
		return fiattokenfactorytypes.QueryGetPauserResponse{}, err
	}

	return showPauserRes, nil
}

// ShowBlacklister queries for the token factory Blacklister by running: `query fiat-tokenfactory show-blacklister`.
func ShowBlacklister(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetBlacklisterResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
	if err != nil {
		return fiattokenfactorytypes.QueryGetBlacklisterResponse{}, err
	}

	var showBlacklisterRes fiattokenfactorytypes.QueryGetBlacklisterResponse
	err = json.Unmarshal(res, &showBlacklisterRes)
	if err != nil {
		return fiattokenfactorytypes.QueryGetBlacklisterResponse{}, err
	}

	return showBlacklisterRes, nil
}

// ShowBlacklisted queries for a specific blacklisted address by running: `query fiat-tokenfactory show-blacklisted <address>`.
// An error is returned if the address is not blacklisted
func ShowBlacklisted(ctx context.Context, val *cosmos.ChainNode, blacklistedWallet ibc.Wallet) (fiattokenfactorytypes.QueryGetBlacklistedResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", blacklistedWallet.FormattedAddress())
	if err != nil {
		return fiattokenfactorytypes.QueryGetBlacklistedResponse{}, err
	}

	var showBlacklistedRes fiattokenfactorytypes.QueryGetBlacklistedResponse
	err = json.Unmarshal(res, &showBlacklistedRes)
	if err != nil {
		return fiattokenfactorytypes.QueryGetBlacklistedResponse{}, err
	}

	return showBlacklistedRes, nil
}

// ShowPaused queries the paused state of the token factory by running: `query fiat-tokenfactory show-paused`.
func ShowPaused(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetPausedResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	if err != nil {
		return fiattokenfactorytypes.QueryGetPausedResponse{}, err
	}

	var showPausedRes fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedRes)
	if err != nil {
		return fiattokenfactorytypes.QueryGetPausedResponse{}, err
	}

	return showPausedRes, nil
}
