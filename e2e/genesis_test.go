package e2e_test

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

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

	denomMetadataUsdc = []banktypes.Metadata{
		{
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
			Base: "uusdc",

			Display: "usdc",
			Name:    "usdc",
			Symbol:  "USDC",
		},
	}
)

type genesisWrapper struct {
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

	// @TODO: do we need these?
	// proposaltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	// upgradetypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func nobleChainSpec(
	ctx context.Context,
	gw *genesisWrapper,
	chainID string,
	nv, nf int,
	minSetupFiatTf bool,
	minModifyGenesisFiatTf bool,
) *interchaintest.ChainSpec {
	return &interchaintest.ChainSpec{
		NumValidators: &nv,
		NumFullNodes:  &nf,
		ChainConfig: ibc.ChainConfig{
			Type:           "cosmos",
			Name:           "noble",
			ChainID:        chainID,
			Bin:            "nobled",
			Denom:          "utoken",
			Bech32Prefix:   "noble",
			CoinType:       "118",
			GasPrices:      "0.0utoken",
			GasAdjustment:  1.1,
			TrustingPeriod: "504h",
			NoHostMount:    false,
			Images:         nobleImageInfo,
			EncodingConfig: NobleEncoding(),
			PreGenesis:     preGenesisAll(ctx, gw, minSetupFiatTf),
			ModifyGenesis:  modifyGenesisAll(gw, minModifyGenesisFiatTf),
		},
	}
}

func modifyGenesisAll(gw *genesisWrapper, minModifyGenesisFiatTf bool) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {

	//@Todo: minModifyGenesisFiatTf

	fiatTokenfactoryGenesis := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.authority.owner", gw.authority.FormattedAddress()),
		cosmos.NewGenesisKV("app_state.bank.denom_metadata", denomMetadataUsdc),
		cosmos.NewGenesisKV("app_state.fiat-tokenfactory.paused", fiattokenfactorytypes.Paused{Paused: false}),
		cosmos.NewGenesisKV("app_state.fiat-tokenfactory.mintingDenom", fiattokenfactorytypes.MintingDenom{Denom: denomMetadataUsdc[0].Base}),
		cosmos.NewGenesisKV("app_state.staking.params.bond_denom", "utoken"),
	}

	return cosmos.ModifyGenesis(fiatTokenfactoryGenesis)

}

func preGenesisAll(ctx context.Context, gw *genesisWrapper, minSetupFiatTf bool) func(ibc.ChainConfig) error {
	return func(cc ibc.ChainConfig) (err error) {
		val := gw.chain.Validators[0]

		gw.fiatTfRoles, err = createTokenfactoryRoles(ctx, denomMetadataUsdc, val, minSetupFiatTf)
		if err != nil {
			return err
		}

		gw.authority, err = createAuthorityRole(ctx, val)
		if err != nil {
			return err
		}

		return err
	}
}

// Creates tokenfactory wallets with 0 amount. Meant to run pre-genesis.
// If minsteup = true, only the owner wallet is created.
// Having minsetup = failse is specifically usefull when wanting the noble roles to be assigned at chain launch.
// After creating thw wallets, it recovers the key on the specified validator.
func createTokenfactoryRoles(ctx context.Context, denomMetadata []banktypes.Metadata, val *cosmos.ChainNode, minSetup bool) (NobleRoles, error) {
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
	if minSetup {
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
