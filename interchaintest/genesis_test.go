package interchaintest_test

import (
	"context"
	"encoding/json"
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
	denomMetadata = []DenomMetadata{
		{
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
		},
		{
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
		},
	}
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
func createTokenfactoryRoles(ctx context.Context, nobleRoles *NobleRoles, val *cosmos.ChainNode) error {
	chainCfg := val.Chain.Config()

	kr := keyring.NewInMemory()

	nobleRoles.Owner = interchaintest.BuildWallet(kr, "owner", chainCfg)
	nobleRoles.Owner2 = interchaintest.BuildWallet(kr, "owner2", chainCfg)
	nobleRoles.MasterMinter = interchaintest.BuildWallet(kr, "masterminter", chainCfg)
	nobleRoles.MinterController = interchaintest.BuildWallet(kr, "mintercontroller", chainCfg)
	nobleRoles.MinterController2 = interchaintest.BuildWallet(kr, "mintercontroller2", chainCfg)
	nobleRoles.Minter = interchaintest.BuildWallet(kr, "minter", chainCfg)
	nobleRoles.Blacklister = interchaintest.BuildWallet(kr, "blacklister", chainCfg)
	nobleRoles.Pauser = interchaintest.BuildWallet(kr, "pauser", chainCfg)

	err := val.RecoverKey(ctx, nobleRoles.Owner.KeyName, nobleRoles.Owner.Mnemonic)
	if err != nil {
		return err
	}
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
			Address: nobleRoles.Owner.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
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
func modifyGenesisTokenfactory(genbz []byte, tokenfactory string, roles *NobleRoles, minSetup bool) ([]byte, error) {
	g := make(map[string]interface{})
	if err := json.Unmarshal(genbz, &g); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Owner.Address}, "app_state", tokenfactory, "owner"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryPaused{false}, "app_state", tokenfactory, "paused"); err != nil {
		return nil, fmt.Errorf("failed to set paused in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryDenom{denomMetadata[0].Base}, "app_state", tokenfactory, "mintingDenom"); err != nil {
		return nil, fmt.Errorf("failed to set minting denom in genesis json: %w", err)
	}
	if minSetup {
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}

	if err := dyno.Set(g, TokenFactoryAddress{roles.MasterMinter.Address}, "app_state", tokenfactory, "masterMinter"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Blacklister.Address}, "app_state", tokenfactory, "blacklister"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{roles.Pauser.Address}, "app_state", tokenfactory, "pauser"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	out, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
	}
	return out, nil
}

func modifyGenesisParamAuthority(genbz []byte, authorityAddress string) ([]byte, error) {
	g := make(map[string]interface{})
	if err := json.Unmarshal(genbz, &g); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
	}
	if err := dyno.Set(g, authorityAddress, "app_state", "params", "params", "authority"); err != nil {
		return nil, fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(g, authorityAddress, "app_state", "upgrade", "params", "authority"); err != nil {
		return nil, fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}
	out, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
	}
	return out, nil
}

func modifyGenesisDenommetadata(genbz []byte) ([]byte, error) {
	g := make(map[string]interface{})
	if err := json.Unmarshal(genbz, &g); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
	}
	if err := dyno.Set(g, denomMetadata, "app_state", "bank", "denom_metadata"); err != nil {
		return nil, fmt.Errorf("failed to set denom metadata in genesis json: %w", err)
	}
	out, err := json.Marshal(g)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
	}
	return out, nil
}
