package ibctest_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/ibctest/v3"
	"github.com/strangelove-ventures/ibctest/v3/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v3/ibc"
	tokenfactorytypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
)

const (
	authorityKeyName = "authority"

	ownerKeyName            = "owner"
	masterMinterKeyName     = "masterminter"
	minterKeyName           = "minter"
	minterControllerKeyName = "mintercontroller"
	blacklisterKeyName      = "blacklister"
	pauserKeyName           = "pauser"
	userKeyName             = "user"
	user2KeyName            = "user2"
	aliceKeyName            = "alice"

	mintingDenom = "urupee"
)

var (
	denomMetadata = []DenomMetadata{
		{
			Display: "rupee",
			Base:    "urupee",
			Name:    "USDC",
			Symbol:  "USDC",
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

type NobleRoles struct {
	Authority        ibc.Wallet
	Owner            ibc.Wallet
	MasterMinter     ibc.Wallet
	MinterController ibc.Wallet
	Minter           ibc.Wallet
	Blacklister      ibc.Wallet
	Pauser           ibc.Wallet
	User             ibc.Wallet
	User2            ibc.Wallet
	Alice            ibc.Wallet
}

func noblePreGenesis(ctx context.Context, val *cosmos.ChainNode) (NobleRoles, error) {
	chainCfg := val.Chain.Config()

	kr := keyring.NewInMemory()

	authority := ibctest.BuildWallet(kr, authorityKeyName, chainCfg)

	masterMinter := ibctest.BuildWallet(kr, masterMinterKeyName, chainCfg)
	minter := ibctest.BuildWallet(kr, minterKeyName, chainCfg)
	owner := ibctest.BuildWallet(kr, ownerKeyName, chainCfg)
	minterController := ibctest.BuildWallet(kr, minterControllerKeyName, chainCfg)
	blacklister := ibctest.BuildWallet(kr, blacklisterKeyName, chainCfg)
	pauser := ibctest.BuildWallet(kr, pauserKeyName, chainCfg)
	user := ibctest.BuildWallet(kr, userKeyName, chainCfg)
	user2 := ibctest.BuildWallet(kr, user2KeyName, chainCfg)
	alice := ibctest.BuildWallet(kr, aliceKeyName, chainCfg)

	err := val.RecoverKey(ctx, authorityKeyName, authority.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, ownerKeyName, owner.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, masterMinterKeyName, masterMinter.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, minterControllerKeyName, minterController.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, minterKeyName, minter.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, blacklisterKeyName, blacklister.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, pauserKeyName, pauser.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, userKeyName, user.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, user2KeyName, user2.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	err = val.RecoverKey(ctx, aliceKeyName, alice.Mnemonic)
	if err != nil {
		return NobleRoles{}, err
	}
	genesisWallets := []ibc.WalletAmount{
		{
			Address: authority.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: owner.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: masterMinter.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: minter.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: minterController.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: blacklister.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: pauser.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: user.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
		{
			Address: user2.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000, // used in testing to test non-tokenfactory assets
		},
		{
			Address: alice.Address,
			Denom:   chainCfg.Denom,
			Amount:  0,
		},
	}

	for _, wallet := range genesisWallets {
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		if err != nil {
			return NobleRoles{}, err
		}
	}
	return NobleRoles{
		Authority:        authority,
		Owner:            owner,
		MasterMinter:     masterMinter,
		MinterController: minterController,
		Minter:           minter,
		Blacklister:      blacklister,
		Pauser:           pauser,
		User:             user,
		User2:            user2,
		Alice:            alice,
	}, nil
}

// Sets the minamum genesis modifications needed to start chain
// Owner account is used for both tokenfactory owner and param authority
func modifyGenesisNobleOwner(genbz []byte, ownerAddress string) ([]byte, error) {
	g := make(map[string]interface{})
	if err := json.Unmarshal(genbz, &g); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{ownerAddress}, "app_state", "tokenfactory", "owner"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, ownerAddress, "app_state", "params", "params", "authority"); err != nil {
		return nil, fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(g, ownerAddress, "app_state", "upgrade", "params", "authority"); err != nil {
		return nil, fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryPaused{false}, "app_state", "tokenfactory", "paused"); err != nil {
		return nil, fmt.Errorf("failed to set paused in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryDenom{mintingDenom}, "app_state", "tokenfactory", "mintingDenom"); err != nil {
		return nil, fmt.Errorf("failed to set minting denom in genesis json: %w", err)
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

// Sets the aurhority, owner, masterminter, blacklister and pauser to separate accounts in genesis.
func modifyGenesisNobleAll(genbz []byte, authorityAddress, ownerAddress, masterMinterAddress, blacklisterAddress, pauserAddress string) ([]byte, error) {
	g := make(map[string]interface{})
	if err := json.Unmarshal(genbz, &g); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{ownerAddress}, "app_state", "tokenfactory", "owner"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{masterMinterAddress}, "app_state", "tokenfactory", "masterMinter"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{blacklisterAddress}, "app_state", "tokenfactory", "blacklister"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{pauserAddress}, "app_state", "tokenfactory", "pauser"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
	}
	if err := dyno.Set(g, authorityAddress, "app_state", "params", "params", "authority"); err != nil {
		return nil, fmt.Errorf("failed to set params authority in genesis json: %w", err)
	}
	if err := dyno.Set(g, authorityAddress, "app_state", "upgrade", "params", "authority"); err != nil {
		return nil, fmt.Errorf("failed to set upgrade authority address in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryPaused{false}, "app_state", "tokenfactory", "paused"); err != nil {
		return nil, fmt.Errorf("failed to set paused in genesis json: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryDenom{mintingDenom}, "app_state", "tokenfactory", "mintingDenom"); err != nil {
		return nil, fmt.Errorf("failed to set minting denom in genesis json: %w", err)
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
