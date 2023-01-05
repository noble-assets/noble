package ibctest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/ibctest/v3"
	"github.com/strangelove-ventures/ibctest/v3/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v3/ibc"
	integration "github.com/strangelove-ventures/noble/ibctest"
	tokenfactorytypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	ownerKeyName            = "owner"
	masterMinterKeyName     = "masterminter"
	minterKeyName           = "minter"
	minterControllerKeyName = "mintercontroller"
	blacklisterKeyName      = "blacklister"
	pauserKeyName           = "pauser"
	userKeyName             = "user"
	user2KeyName            = "user2"
	aliceKeyName            = "alice"

	mintingDenom = "uusdc"
)

var (
	denomMetadata = []DenomMetadata{
		{
			Display: "usdc",
			Base:    "uusdc",
			Name:    "USDC",
			Symbol:  "USDC",
			DenomUnits: []DenomUnit{
				{
					Denom: "uusdc",
					Aliases: []string{
						"microusdc",
					},
					Exponent: "0",
				},
				{
					Denom: "musdc",
					Aliases: []string{
						"milliusdc",
					},
					Exponent: "3",
				},
				{
					Denom:    "usdc",
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

func HeroEncoding() *simappparams.EncodingConfig {
	cfg := cosmos.DefaultEncoding()

	// register custom types
	tokenfactorytypes.RegisterInterfaces(cfg.InterfaceRegistry)

	return &cfg
}

func TestHeroChain(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	client, network := ibctest.DockerSetup(t)

	repo, version := integration.GetDockerImageInfo()

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "token",
		Bech32Prefix:   "cosmos",
		GasPrices:      "0.0token",
		GasAdjustment:  1.1,
		TrustingPeriod: "504h",
		NoHostMount:    false,
		Images: []ibc.DockerImage{
			{
				Repository: repo,
				Version:    version,
				UidGid:     "1025:1025",
			},
		},
		EncodingConfig: HeroEncoding(),
	}

	nv := 1
	nf := 0

	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*ibctest.ChainSpec{
		{
			ChainConfig:   chainCfg,
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	noble := chains[0].(*cosmos.CosmosChain)

	err = noble.Initialize(ctx, t.Name(), client, network)
	require.NoError(t, err, "failed to initialize noble chain")

	kr := keyring.NewInMemory()

	masterMinter := ibctest.BuildWallet(kr, masterMinterKeyName, chainCfg)
	minter := ibctest.BuildWallet(kr, minterKeyName, chainCfg)
	owner := ibctest.BuildWallet(kr, ownerKeyName, chainCfg)
	minterController := ibctest.BuildWallet(kr, minterControllerKeyName, chainCfg)
	blacklister := ibctest.BuildWallet(kr, blacklisterKeyName, chainCfg)
	pauser := ibctest.BuildWallet(kr, pauserKeyName, chainCfg)
	user := ibctest.BuildWallet(kr, userKeyName, chainCfg)
	user2 := ibctest.BuildWallet(kr, user2KeyName, chainCfg)
	alice := ibctest.BuildWallet(kr, aliceKeyName, chainCfg)

	nobleValidator := noble.Validators[0]

	err = nobleValidator.RecoverKey(ctx, ownerKeyName, owner.Mnemonic)
	require.NoError(t, err, "failed to restore owner key")

	err = nobleValidator.RecoverKey(ctx, masterMinterKeyName, masterMinter.Mnemonic)
	require.NoError(t, err, "failed to restore masterminter key")

	err = nobleValidator.RecoverKey(ctx, minterControllerKeyName, minterController.Mnemonic)
	require.NoError(t, err, "failed to restore mintercontroller key")

	err = nobleValidator.RecoverKey(ctx, minterKeyName, minter.Mnemonic)
	require.NoError(t, err, "failed to restore minter key")

	err = nobleValidator.RecoverKey(ctx, blacklisterKeyName, blacklister.Mnemonic)
	require.NoError(t, err, "failed to restore blacklister key")

	err = nobleValidator.RecoverKey(ctx, pauserKeyName, pauser.Mnemonic)
	require.NoError(t, err, "failed to restore pauser key")

	err = nobleValidator.RecoverKey(ctx, userKeyName, user.Mnemonic)
	require.NoError(t, err, "failed to restore user key")

	err = nobleValidator.RecoverKey(ctx, user2KeyName, user2.Mnemonic)
	require.NoError(t, err, "failed to restore user key")

	err = nobleValidator.RecoverKey(ctx, aliceKeyName, alice.Mnemonic)
	require.NoError(t, err, "failed to restore alice key")

	err = nobleValidator.InitFullNodeFiles(ctx)
	require.NoError(t, err, "failed to initialize noble validator config")

	genesisWallets := []ibc.WalletAmount{
		{
			Address: owner.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: masterMinter.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: minter.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: minterController.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: blacklister.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: pauser.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: user.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: user2.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
		{
			Address: alice.Address,
			Denom:   chainCfg.Denom,
			Amount:  10_000,
		},
	}

	for _, wallet := range genesisWallets {
		err = nobleValidator.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		require.NoError(t, err, "failed to add genesis account")
	}

	genBz, err := nobleValidator.GenesisFileContent(ctx)
	require.NoError(t, err, "failed to read genesis file")

	genBz, err = modifyGenesisHero(genBz, owner.Address)
	require.NoError(t, err, "failed to modify genesis file")

	err = nobleValidator.OverwriteGenesisFile(ctx, genBz)
	require.NoError(t, err, "failed to write genesis file")

	_, _, err = nobleValidator.ExecBin(ctx, "add-consumer-section")
	require.NoError(t, err, "failed to add consumer section to noble validator genesis file")

	err = nobleValidator.CreateNodeContainer(ctx)
	require.NoError(t, err, "failed to create noble validator container")

	err = nobleValidator.StartContainer(ctx)
	require.NoError(t, err, "failed to create noble validator container")

	_, err = nobleValidator.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-master-minter", masterMinter.Address,
	)
	require.NoError(t, err, "failed to execute update master minter tx")

	_, err = nobleValidator.ExecTx(ctx, masterMinterKeyName,
		"tokenfactory", "configure-minter-controller", minterController.Address, minter.Address,
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, minterControllerKeyName,
		"tokenfactory", "configure-minter", minter.Address, "1000uusdc",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", user.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(100), userBalance, "failed to mint uusdc to user")

	_, err = nobleValidator.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-blacklister", blacklister.Address,
	)
	require.NoError(t, err, "failed to set blacklister")

	_, err = nobleValidator.ExecTx(ctx, blacklisterKeyName,
		"tokenfactory", "blacklist", user.Address,
	)
	require.NoError(t, err, "failed to blacklist user address")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", user.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", user2.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user2 tx")

	err = nobleValidator.SendFunds(ctx, user2KeyName, ibc.WalletAmount{
		Address: user.Address,
		Denom:   "uusdc",
		Amount:  50,
	})
	require.Error(t, err, "The tx to a blacklisted user should not have been successful")

	userBalance, err = noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(100), userBalance, "user balance should not have incremented while blacklisted")

	err = nobleValidator.SendFunds(ctx, user2KeyName, ibc.WalletAmount{
		Address: user.Address,
		Denom:   "token",
		Amount:  100,
	})
	require.NoError(t, err, "The tx should have been successfull as that is no the minting denom")

	userBalance, err = noble.GetBalance(ctx, user.Address, "token")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(10_100), userBalance, "user balance should have incremented")

	_, err = nobleValidator.ExecTx(ctx, blacklisterKeyName,
		"tokenfactory", "unblacklist", user.Address,
	)
	require.NoError(t, err, "failed to unblacklist user address")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", user.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should have increased now that they are no longer blacklisted")

	_, err = nobleValidator.ExecTx(ctx, ownerKeyName,
		"tokenfactory", "update-pauser", pauser.Address,
	)
	require.NoError(t, err, "failed to update pauser")

	_, err = nobleValidator.ExecTx(ctx, pauserKeyName,
		"tokenfactory", "pause",
	)
	require.NoError(t, err, "failed to pause mints")

	_, err = nobleValidator.ExecTx(ctx, minterKeyName,
		"tokenfactory", "mint", user.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err = noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, userKeyName,
		"bank", "send", user.Address, alice.Address, "100uusdc",
	)
	require.Error(t, err, "transaction was successful while chain was paused")

	userBalance, err = noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(200), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err := noble.GetBalance(ctx, alice.Address, "uusdc")
	require.NoError(t, err, "failed to get alice balance")

	require.Equal(t, int64(0), aliceBalance, "alice balance should not have increased while chain is paused")

	_, err = nobleValidator.ExecTx(ctx, pauserKeyName,
		"tokenfactory", "unpause",
	)
	require.NoError(t, err, "failed to unpause mints")

	_, err = nobleValidator.ExecTx(ctx, userKeyName,
		"bank", "send", user.Address, alice.Address, "100uusdc",
	)
	require.NoError(t, err, "failed to send tx bank from user to alice")

	userBalance, err = noble.GetBalance(ctx, user.Address, "uusdc")
	require.NoError(t, err, "failed to get user balance")

	require.Equal(t, int64(100), userBalance, "user balance should not have changed while chain is paused")

	aliceBalance, err = noble.GetBalance(ctx, alice.Address, "uusdc")
	require.NoError(t, err, "failed to get alice balance")

	require.Equal(t, int64(100), aliceBalance, "alice balance should not have increased while chain is paused")

}

func modifyGenesisHero(genbz []byte, ownerAddress string) ([]byte, error) {
	g := make(map[string]interface{})
	if err := json.Unmarshal(genbz, &g); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
	}
	if err := dyno.Set(g, TokenFactoryAddress{ownerAddress}, "app_state", "tokenfactory", "owner"); err != nil {
		return nil, fmt.Errorf("failed to set owner address in genesis json: %w", err)
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
