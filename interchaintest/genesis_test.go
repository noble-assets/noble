package interchaintest_test

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	tokenfactorytypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
)

var (
	DenomMetadata_rupee = DenomMetadata{

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
	DenomMetadata_drachma = DenomMetadata{
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
	defaultTransferFeeDenom        = DenomMetadata_drachma.Base
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

type Authority struct {
	Authority ibc.Wallet
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

	kr := keyring.NewInMemory()

	nobleRoles.Owner = interchaintest.BuildWallet(kr, "owner-"+denomMetadata.Base, chainCfg)
	err := val.RecoverKey(ctx, nobleRoles.Owner.KeyName, nobleRoles.Owner.Mnemonic)
	if err != nil {
		return err
	}
	genesisWallet := ibc.WalletAmount{
		Address: nobleRoles.Owner.Address,
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

	nobleRoles.Owner2 = interchaintest.BuildWallet(kr, "owner2-"+denomMetadata.Base, chainCfg)
	nobleRoles.MasterMinter = interchaintest.BuildWallet(kr, "masterminter-"+denomMetadata.Base, chainCfg)
	nobleRoles.MinterController = interchaintest.BuildWallet(kr, "mintercontroller-"+denomMetadata.Base, chainCfg)
	nobleRoles.MinterController2 = interchaintest.BuildWallet(kr, "mintercontroller2-"+denomMetadata.Base, chainCfg)
	nobleRoles.Minter = interchaintest.BuildWallet(kr, "minter-"+denomMetadata.Base, chainCfg)
	nobleRoles.Blacklister = interchaintest.BuildWallet(kr, "blacklister-"+denomMetadata.Base, chainCfg)
	nobleRoles.Pauser = interchaintest.BuildWallet(kr, "pauser-"+denomMetadata.Base, chainCfg)

	err = val.RecoverKey(ctx, nobleRoles.Owner2.KeyName, nobleRoles.Owner2.Mnemonic)
	if err != nil {
		return err
	}
	err = val.RecoverKey(ctx, nobleRoles.MasterMinter.KeyName, nobleRoles.MasterMinter.Mnemonic)
	if err != nil {
		return err
	}
	err = val.RecoverKey(ctx, nobleRoles.MinterController.KeyName, nobleRoles.MinterController.Mnemonic)
	if err != nil {
		return err
	}
	err = val.RecoverKey(ctx, nobleRoles.MinterController2.KeyName, nobleRoles.MinterController2.Mnemonic)
	if err != nil {
		return err
	}
	err = val.RecoverKey(ctx, nobleRoles.Minter.KeyName, nobleRoles.Minter.Mnemonic)
	if err != nil {
		return err
	}
	err = val.RecoverKey(ctx, nobleRoles.Blacklister.KeyName, nobleRoles.Blacklister.Mnemonic)
	if err != nil {
		return err
	}
	err = val.RecoverKey(ctx, nobleRoles.Pauser.KeyName, nobleRoles.Pauser.Mnemonic)
	if err != nil {
		return err
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: nobleRoles.Owner2.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MasterMinter.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MinterController.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.MinterController2.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Minter.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Blacklister.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: nobleRoles.Pauser.Address,
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
func createParamAuthAtGenesis(ctx context.Context, val *cosmos.ChainNode) (Authority, error) {
	chainCfg := val.Chain.Config()

	kr := keyring.NewInMemory()

	authority := &Authority{}

	authority.Authority = interchaintest.BuildWallet(kr, "authority", chainCfg)

	err := val.RecoverKey(ctx, authority.Authority.KeyName, authority.Authority.Mnemonic)
	if err != nil {
		return Authority{}, err
	}

	genesisWallet := ibc.WalletAmount{
		Address: authority.Authority.Address,
		Denom:   chainCfg.Denom,
		Amount:  0,
	}

	err = val.AddGenesisAccount(ctx, genesisWallet.Address, []types.Coin{types.NewCoin(genesisWallet.Denom, types.NewIntFromUint64(uint64(genesisWallet.Amount)))})
	if err != nil {
		return Authority{}, err
	}
	return *authority, nil
}

// Creates extra wallets used for testing. Meant to run pre-genesis.
// It then recovers the key on the specified validator.
func createExtraWalletsAtGenesis(ctx context.Context, val *cosmos.ChainNode) (ExtraWallets, error) {
	chainCfg := val.Chain.Config()

	kr := keyring.NewInMemory()

	extraWallets := &ExtraWallets{}

	extraWallets.User = interchaintest.BuildWallet(kr, "user", chainCfg)
	extraWallets.User2 = interchaintest.BuildWallet(kr, "user2", chainCfg)
	extraWallets.Alice = interchaintest.BuildWallet(kr, "alice", chainCfg)

	err := val.RecoverKey(ctx, extraWallets.User.KeyName, extraWallets.User.Mnemonic)
	if err != nil {
		return ExtraWallets{}, err
	}
	err = val.RecoverKey(ctx, extraWallets.User2.KeyName, extraWallets.User2.Mnemonic)
	if err != nil {
		return ExtraWallets{}, err
	}
	err = val.RecoverKey(ctx, extraWallets.Alice.KeyName, extraWallets.Alice.Mnemonic)
	if err != nil {
		return ExtraWallets{}, err
	}

	genesisWallets := []ibc.WalletAmount{
		{
			Address: extraWallets.User.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: extraWallets.User2.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: extraWallets.Alice.Address,
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
	if err := dyno.Set(g, TokenFactoryAddress{roles.Owner.Address}, "app_state", tokenfactoryModName, "owner"); err != nil {
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
	if err := dyno.Set(g, TokenFactoryAddress{roles.MasterMinter.Address}, "app_state", tokenfactoryModName, "masterMinter"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Blacklister.Address}, "app_state", tokenfactoryModName, "blacklister"); err != nil {
		return fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Pauser.Address}, "app_state", tokenfactoryModName, "pauser"); err != nil {
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
