package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"

	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
)

var (
	nobleImageInfo = []ibc.DockerImage{
		{
			Repository: "noble",
			Version:    "local",
			UidGid:     "1025:1025",
		},
	}

	denomMetadataUsdc = banktypes.Metadata{
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

type nobleWrapper struct {
	chain       *cosmos.CosmosChain
	fiatTfRoles NobleRoles
	authority   ibc.Wallet
}

type NobleRoles struct {
	Owner            ibc.Wallet
	MasterMinter     ibc.Wallet
	MinterController ibc.Wallet
	Minter           ibc.Wallet
	Blacklister      ibc.Wallet
	Pauser           ibc.Wallet
}

func NobleEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	fiattokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

func nobleChainSpec(
	ctx context.Context,
	gw *nobleWrapper,
	chainID string,
	nv, nf int,
	setupAllFiatTFRoles bool,
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
			Images:         nobleImageInfo,
			EncodingConfig: NobleEncoding(),
			PreGenesis:     preGenesisAll(ctx, gw, setupAllFiatTFRoles),
			ModifyGenesis:  modifyGenesisAll(gw, setupAllFiatTFRoles),
		},
	}
}

// modifyGenesisAll modifies the genesis file to with fields needed to start chain
//
// setupAllFiatTFRoles: if true, all Tokenfactory roles will be setup and tied to a wallet at genesis,
// if false, only the Owner role will be setup
func modifyGenesisAll(nw *nobleWrapper, setupAllFiatTFRoles bool) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {

		updatedGenesis := []cosmos.GenesisKV{
			cosmos.NewGenesisKV("app_state.authority.owner", nw.authority.FormattedAddress()),
			cosmos.NewGenesisKV("app_state.bank.denom_metadata", []banktypes.Metadata{denomMetadataUsdc}),
			cosmos.NewGenesisKV("app_state.fiat-tokenfactory.owner", fiattokenfactorytypes.Owner{Address: nw.fiatTfRoles.Owner.FormattedAddress()}),
			cosmos.NewGenesisKV("app_state.fiat-tokenfactory.paused", fiattokenfactorytypes.Paused{Paused: false}),
			cosmos.NewGenesisKV("app_state.fiat-tokenfactory.mintingDenom", fiattokenfactorytypes.MintingDenom{Denom: denomMetadataUsdc.Base}),
			cosmos.NewGenesisKV("app_state.staking.params.bond_denom", "ustake"),
		}

		if setupAllFiatTFRoles {
			allFiatTFRoles := []cosmos.GenesisKV{
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.masterMinter", fiattokenfactorytypes.MasterMinter{Address: nw.fiatTfRoles.MasterMinter.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.mintersList", []fiattokenfactorytypes.Minters{{Address: nw.fiatTfRoles.Minter.FormattedAddress(), Allowance: sdktypes.Coin{Denom: denomMetadataUsdc.Base, Amount: math.NewInt(100_00_000)}}}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.pauser", fiattokenfactorytypes.Pauser{Address: nw.fiatTfRoles.Pauser.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.blacklister", fiattokenfactorytypes.Blacklister{Address: nw.fiatTfRoles.Blacklister.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.masterMinter", fiattokenfactorytypes.MasterMinter{Address: nw.fiatTfRoles.MasterMinter.FormattedAddress()}),
				cosmos.NewGenesisKV("app_state.fiat-tokenfactory.minterControllerList", []fiattokenfactorytypes.MinterController{{Minter: nw.fiatTfRoles.Minter.FormattedAddress(), Controller: nw.fiatTfRoles.MinterController.FormattedAddress()}}),
			}
			updatedGenesis = append(updatedGenesis, allFiatTFRoles...)
		}

		return cosmos.ModifyGenesis(updatedGenesis)(cc, b)
	}
}

func preGenesisAll(ctx context.Context, nw *nobleWrapper, setupAllFiatTFRoles bool) func(ibc.ChainConfig) error {
	return func(cc ibc.ChainConfig) (err error) {
		val := nw.chain.Validators[0]

		nw.fiatTfRoles, err = createTokenfactoryRoles(ctx, val, setupAllFiatTFRoles)
		if err != nil {
			return err
		}

		nw.authority, err = createAuthorityRole(ctx, val)
		if err != nil {
			return err
		}

		return err
	}
}

// createTokenfactoryRoles Creates wallets to be tied to TF roles with 0 amount. Meant to run pre-genesis.
// After creating thw wallets, it recovers the key on the specified validator.
//
// setupAllFiatTFRoles: if true, a wallet for all Tokenfactory roles will be created,
// if false, only the Owner role will be created
func createTokenfactoryRoles(ctx context.Context, val *cosmos.ChainNode, setupAllFiatTFRoles bool) (NobleRoles, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error
	nobleRoles := NobleRoles{}

	nobleRoles.Owner, err = nobleVal.BuildRelayerWallet(ctx, "owner-fiatTF")
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	if err := val.RecoverKey(ctx, nobleRoles.Owner.KeyName(), nobleRoles.Owner.Mnemonic()); err != nil {
		return NobleRoles{}, fmt.Errorf("failed to restore %s wallet: %w", nobleRoles.Owner.KeyName(), err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: nobleRoles.Owner.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  math.ZeroInt(),
	}
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []sdktypes.Coin{sdktypes.NewCoin(genesisWallet.Denom, genesisWallet.Amount)})
	if err != nil {
		return NobleRoles{}, err
	}

	if !setupAllFiatTFRoles {
		return nobleRoles, nil
	}

	nobleRoles.MasterMinter, err = nobleVal.BuildRelayerWallet(ctx, "masterminter")
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "masterminter", err)
	}
	nobleRoles.MinterController, err = nobleVal.BuildRelayerWallet(ctx, "mintercontroller")
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "mintercontroller", err)
	}
	nobleRoles.Minter, err = nobleVal.BuildRelayerWallet(ctx, "minter")
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "minter", err)
	}
	nobleRoles.Blacklister, err = nobleVal.BuildRelayerWallet(ctx, "blacklister")
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "blacklister", err)
	}
	nobleRoles.Pauser, err = nobleVal.BuildRelayerWallet(ctx, "pauser")
	if err != nil {
		return NobleRoles{}, fmt.Errorf("failed to create %s wallet: %w", "pauser", err)
	}

	walletsToRestore := []ibc.Wallet{nobleRoles.MasterMinter, nobleRoles.MinterController, nobleRoles.Minter, nobleRoles.Blacklister, nobleRoles.Pauser}
	for _, wallet := range walletsToRestore {
		if err = val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return NobleRoles{}, fmt.Errorf("failed to restore %s wallet: %w", wallet.KeyName(), err)
		}
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: nobleRoles.MasterMinter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.MinterController.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.Minter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.Blacklister.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
		{
			Address: nobleRoles.Pauser.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  math.ZeroInt(),
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []sdktypes.Coin{sdktypes.NewCoin(wallet.Denom, wallet.Amount)})
		if err != nil {
			return NobleRoles{}, err
		}
	}

	return nobleRoles, nil
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
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []sdktypes.Coin{sdktypes.NewCoin(genesisWallet.Denom, genesisWallet.Amount)})
	if err != nil {
		return nil, err
	}

	return authority, nil
}

// nobleSpinUp starts noble chain
//
// setupAllFiatTFRoles: if true, all Tokenfactory roles will be created and setup at genesis,
// if false, only the Owner role will be created
func nobleSpinUp(t *testing.T, ctx context.Context, setupAllFiatTFRoles bool) (nw nobleWrapper) {
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	numValidators := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &nw, "noble-1", numValidators, numFullNodes, setupAllFiatTFRoles),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	nw.chain = chains[0].(*cosmos.CosmosChain)
	noble := nw.chain

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

// nobleSpinUpIBC is the same as nobleSpinUp but it also spins up a gaia chain and creates
// an IBC path between them
//
// setupAllFiatTFRoles: if true, all Tokenfactory roles will be created and setup at genesis,
// if false, only the Owner role will be created
func nobleSpinUpIBC(t *testing.T, ctx context.Context, setupAllFiatTFRoles bool) (nw nobleWrapper, gaia *cosmos.CosmosChain, r ibc.Relayer, ibcPathName string, eRep *testreporter.RelayerExecReporter) {
	rep := testreporter.NewNopReporter()
	eRep = rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	numValidators := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &nw, "noble-1", numValidators, numFullNodes, setupAllFiatTFRoles),
		{Name: "gaia", Version: "v16.0.0", NumValidators: &numValidators, NumFullNodes: &numFullNodes},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	nw.chain = chains[0].(*cosmos.CosmosChain)
	noble := nw.chain
	gaia = chains[1].(*cosmos.CosmosChain)

	rf := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t))
	r = rf.Build(t, client, network)

	ibcPathName = "path"
	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: r,
			Path:    ibcPathName,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

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

// blacklistAccount blacklists an account and then runs the `show-blacklisted` query to ensure the
// account was successfully blacklisted on chain
func blacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, toBlacklist ibc.Wallet) {
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

// unblacklistAccount unblacklists an account and then runs the `show-blacklisted` query to ensure the
// account was successfully unblacklisted on chain
func unblacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, unBlacklist ibc.Wallet) {
	_, err := val.ExecTx(ctx, blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", unBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", unBlacklist.FormattedAddress())
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
}

// pauseFiatTF pauses the fiat tokenfactory. It then runs the `show-paused` query to ensure the
// the tokenfactory was successfully paused
func pauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser ibc.Wallet) {
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

// unpauseFiatTF pauses the fiat tokenfactory. It then runs the `show-paused` query to ensure the
// the tokenfactory was successfully unpaused
func unpauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser ibc.Wallet) {
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

// setupMinterAndController creates a minter controller and minter. It also sets up a minter with an specified allowance of `uusdc`
func setupMinterAndController(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, val *cosmos.ChainNode, masterMinter ibc.Wallet, allowance int64) (minter ibc.Wallet, minterController ibc.Wallet) {
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	minterController = w[0]
	minter = w[1]

	_, err := val.ExecTx(ctx, masterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController.FormattedAddress(), minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMC, err := showMinterController(ctx, val, minterController)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter.FormattedAddress(),
			Controller: minterController.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMC.MinterController)

	configureMinter(t, ctx, val, minterController, minter, allowance)

	return minter, minterController
}

// configureMinter configures a minter with a specified allowance of `uusdc`. It then runs the `show-minters` query to ensure
// the minter was properly configured
func configureMinter(t *testing.T, ctx context.Context, val *cosmos.ChainNode, minterController, minter ibc.Wallet, allowance int64) {
	_, err := val.ExecTx(ctx, minterController.KeyName(), "fiat-tokenfactory", "configure-minter", minter.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.NoError(t, err, "error configuring minter")

	showMinter, err := showMinters(ctx, val, minter)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinter.Minters)
}

// showMinterController queries for a specific minter controller by running: `query fiat-tokenfactory show-minter-controller <address>`.
// An error is returned if the minter controller does not exist
func showMinterController(ctx context.Context, val *cosmos.ChainNode, minterController ibc.Wallet) (fiattokenfactorytypes.QueryGetMinterControllerResponse, error) {
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

// showMinters queries for a specific minter by running: `query fiat-tokenfactory show-minters <address>`.
// An error is returned if the minter does not exist
func showMinters(ctx context.Context, val *cosmos.ChainNode, minter ibc.Wallet) (fiattokenfactorytypes.QueryGetMintersResponse, error) {
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

// showOwner queries for the token factory Owner by running: `query fiat-tokenfactory show-owner`.
func showOwner(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetOwnerResponse, error) {
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

// showMasterMinter queries for the token factory Master Minter by running: `query fiat-tokenfactory show-master-minter`.
func showMasterMinter(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetMasterMinterResponse, error) {
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

// showPauser queries for the token factory Pauser by running: `query fiat-tokenfactory show-pauser`.
func showPauser(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetPauserResponse, error) {
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

// showBlacklister queries for the token factory Blacklister by running: `query fiat-tokenfactory show-blacklister`.
func showBlacklister(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetBlacklisterResponse, error) {
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

// showBlacklisted queries for a specific blacklisted address by running: `query fiat-tokenfactory show-blacklisted <address>`.
// An error is returned if the address is not blacklisted
func showBlacklisted(ctx context.Context, val *cosmos.ChainNode, blacklistedWallet ibc.Wallet) (fiattokenfactorytypes.QueryGetBlacklistedResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", blacklistedWallet.FormattedAddress())
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

// showPaused queries paused state the token factory by running: `query fiat-tokenfactory show-paused`.
func showPaused(ctx context.Context, val *cosmos.ChainNode) (fiattokenfactorytypes.QueryGetPausedResponse, error) {
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
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
