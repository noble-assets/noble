package interchaintest_test

import (
	"context"
	"fmt"

	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/relayer"
	"github.com/strangelove-ventures/interchaintest/v3/relayer/rly"
	tokenfactorytypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
)

var (
	denomMetadataFrienzies = DenomMetadata{
		Display: "ufrienzies",
		Base:    "ufrienzies",
		Name:    "frienzies",
		Symbol:  "FRNZ",
		DenomUnits: []DenomUnit{
			{
				Denom: "ufrienzies",
				Aliases: []string{
					"microfrienzies",
				},
				Exponent: "0",
			},
			{
				Denom: "mfrienzies",
				Aliases: []string{
					"millifrienzies",
				},
				Exponent: "3",
			},
			{
				Denom:    "frienzies",
				Exponent: "6",
			},
		},
	}

	denomMetadataRupee = DenomMetadata{
		Display: "rupee",
		Base:    "urupee",
		Name:    "rupee",
		Symbol:  "RUPEE",
		DenomUnits: []DenomUnit{
			{
				Denom: "urupee",
				Aliases: []string{
					"microrupee",
				},
				Exponent: "0",
			},
			{
				Denom: "mrupee",
				Aliases: []string{
					"millirupee",
				},
				Exponent: "3",
			},
			{
				Denom:    "rupee",
				Exponent: "6",
			},
		},
	}

	denomMetadataDrachma = DenomMetadata{
		Display: "drachma",
		Base:    "udrachma",
		Name:    "drachma",
		Symbol:  "DRACHMA",
		DenomUnits: []DenomUnit{
			{
				Denom: "udrachma",
				Aliases: []string{
					"microdrachma",
				},
				Exponent: "0",
			},
			{
				Denom: "mdrachma",
				Aliases: []string{
					"millidrachma",
				},
				Exponent: "3",
			},
			{
				Denom:    "drachma",
				Exponent: "6",
			},
		},
	}

	defaultShare                   = "0.8"
	defaultDistributionEntityShare = "1.0"
	defaultTransferBPSFee          = "1"
	defaultTransferMaxFee          = "5000000"
	defaultTransferFeeDenom        = denomMetadataDrachma.Base

	relayerImage = relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "v2.3.1", rly.RlyDefaultUidGid)
)

type DenomMetadata struct {
	Display    string      `json:"display"`
	Base       string      `json:"base"`
	Name       string      `json:"name"`
	Symbol     string      `json:"symbol"`
	DenomUnits []DenomUnit `json:"denom_units"`
}

type DenomUnit struct {
	Denom    string   `json:"denom"`
	Aliases  []string `json:"aliases"`
	Exponent string   `json:"exponent"`
}

type TokenFactoryAddress struct {
	Address string `json:"address"`
}

type ParamAuthAddress struct {
	Address string `json:"address"`
}

type TokenFactoryPaused struct {
	Paused bool `json:"paused"`
}

type TokenFactoryDenom struct {
	Denom string `json:"denom"`
}

type DistributionEntity struct {
	Address string `json:"address"`
	Share   string `json:"share"`
}

func NobleEncoding() *simappparams.EncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)
	proposaltypes.RegisterInterfaces(cfg.InterfaceRegistry)
	upgradetypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

type ExtraWallets struct {
	User  ibc.Wallet
	User2 ibc.Wallet
	Alice ibc.Wallet
}

type NobleRoles struct {
	Owner             ibc.Wallet
	Owner2            ibc.Wallet
	MasterMinter      ibc.Wallet
	MinterController  ibc.Wallet
	MinterController2 ibc.Wallet
	Minter            ibc.Wallet
	Blacklister       ibc.Wallet
	Pauser            ibc.Wallet
}

// Creates tokenfactory wallets. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createTokenfactoryRoles(ctx context.Context, nobleRoles *NobleRoles, denomMetadata DenomMetadata, val *cosmos.ChainNode, minSetup bool) error {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	nobleRoles.Owner, err = nobleVal.BuildWallet(ctx, "owner-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: string(nobleRoles.Owner.FormattedAddress()),
		Denom:   chainCfg.Denom,
		Amount:  0,
	}
	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, types.NewIntFromUint64(uint64(genesisWallet.Amount)))})
	if err != nil {
		return err
	}
	if minSetup {
		return nil
	}

	nobleRoles.Owner2, err = nobleVal.BuildWallet(ctx, "owner2-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	nobleRoles.MasterMinter, err = nobleVal.BuildWallet(ctx, "masterminter-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	nobleRoles.MinterController, err = nobleVal.BuildWallet(ctx, "mintercontroller-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	nobleRoles.MinterController2, err = nobleVal.BuildWallet(ctx, "mintercontroller2-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	nobleRoles.Minter, err = nobleVal.BuildWallet(ctx, "minter-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	nobleRoles.Blacklister, err = nobleVal.BuildWallet(ctx, "blacklister-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}
	nobleRoles.Pauser, err = nobleVal.BuildWallet(ctx, "pauser-"+denomMetadata.Base, "")
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: nobleRoles.Owner2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MasterMinter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MinterController.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MinterController2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Minter.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Blacklister.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Pauser.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		if err != nil {
			return err
		}
	}
	return nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createParamAuthAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ibc.Wallet, error) {
	chainCfg := val.Chain.Config()

	wallet, err := val.Chain.BuildWallet(ctx, "authority", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	genesisWallet := ibc.WalletAmount{
		Address: wallet.FormattedAddress(),
		Denom:   chainCfg.Denom,
		Amount:  0,
	}

	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, types.NewIntFromUint64(uint64(genesisWallet.Amount)))})
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createExtraWalletsAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ExtraWallets, error) {
	chainCfg := val.Chain.Config()
	nobleVal := val.Chain

	var err error

	extraWallets := &ExtraWallets{}

	extraWallets.User, err = nobleVal.BuildWallet(ctx, "user", "")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	extraWallets.User2, err = nobleVal.BuildWallet(ctx, "user2", "")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}
	extraWallets.Alice, err = nobleVal.BuildWallet(ctx, "alice", "")
	if err != nil {
		return ExtraWallets{}, fmt.Errorf("failed to create wallet: %w", err)
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: extraWallets.User.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: extraWallets.User2.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: extraWallets.Alice.FormattedAddress(),
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		if err != nil {
			return ExtraWallets{}, err
		}
	}
	return *extraWallets, nil
}

// Modifies tokenfactory genesis accounts.
// If minSetup = true, only the owner address, paused state, and denom is setup in genesis.
// These are minimum requirements to start the chain. Otherwise all tokenfactory accounts are created.
func modifyGenesisTokenfactory(g map[string]interface{}, tokenfactoryModName string, denomMetadata DenomMetadata, roles *NobleRoles, minSetup bool) error {
	if err := dyno.Set(g, TokenFactoryAddress{roles.Owner.FormattedAddress()}, "app_state", tokenfactoryModName, "owner"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryPaused{false}, "app_state", tokenfactoryModName, "paused"); err != nil {
		return fmt.Errorf("failed to set paused in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryDenom{denomMetadata.Base}, "app_state", tokenfactoryModName, "mintingDenom"); err != nil {
		return fmt.Errorf("failed to set minting denom in genesis json: %w", err)
	}
	if err := dyno.Append(g, denomMetadata, "app_state", "bank", "denom_metadata"); err != nil {
		return fmt.Errorf("failed to set denom metadata in genesis json: %w", err)
	}
	if minSetup {
		return nil
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.MasterMinter.FormattedAddress()}, "app_state", tokenfactoryModName, "masterMinter"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Blacklister.FormattedAddress()}, "app_state", tokenfactoryModName, "blacklister"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Pauser.FormattedAddress()}, "app_state", tokenfactoryModName, "pauser"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	return nil
}

func modifyGenesisParamAuthority(genbz map[string]interface{}, authorityAddress string) error {
	if err := dyno.Set(genbz, authorityAddress, "app_state", "params", "params", "authority"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, authorityAddress, "app_state", "upgrade", "params", "authority"); err != nil {
		return fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}
	return nil
}

func modifyGenesisTariffDefaults(
	genbz map[string]interface{},
	distributionEntity string,
) error {
	return modifyGenesisTariff(genbz, defaultShare, distributionEntity,
		defaultDistributionEntityShare, defaultTransferBPSFee, defaultTransferMaxFee, defaultTransferFeeDenom)
}

func modifyGenesisTariff(
	genbz map[string]interface{},
	share string,
	distributionEntity string,
	distributionEntityShare string,
	transferBPSFee string,
	transferMaxFee string,
	transferDenom string,
) error {
	if err := dyno.Set(genbz, share, "app_state", "tariff", "params", "share"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	distributionEntities := []DistributionEntity{
		{
			Address: distributionEntity,
			Share:   distributionEntityShare,
		},
	}
	if err := dyno.Set(genbz, distributionEntities, "app_state", "tariff", "params", "distribution_entities"); err != nil {
		return fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, transferBPSFee, "app_state", "tariff", "params", "transfer_fee_bps"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, transferMaxFee, "app_state", "tariff", "params", "transfer_fee_max"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(genbz, transferDenom, "app_state", "tariff", "params", "transfer_fee_denom"); err != nil {
		return fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	return nil
}
