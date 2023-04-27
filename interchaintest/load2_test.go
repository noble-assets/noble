package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdkbanktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/noble/cmd"
	integration "github.com/strangelove-ventures/noble/interchaintest"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
)

func TestLoad2(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	repo, version := integration.GetDockerImageInfo()

	var noble *cosmos.CosmosChain
	var roles NobleRoles
	var roles2 NobleRoles
	var paramauthorityWallet Authority

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "utoken",
		Bech32Prefix:   "noble",
		CoinType:       "118",
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
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) error {
			val := noble.Validators[0]
			err := createTokenfactoryRoles(ctx, &roles, DenomMetadata_rupee, val, true)
			if err != nil {
				return err
			}
			err = createTokenfactoryRoles(ctx, &roles2, DenomMetadata_drachma, val, true)
			if err != nil {
				return err
			}
			paramauthorityWallet, err = createParamAuthAtGenesis(ctx, val)
			return err
		},
		ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			g := make(map[string]interface{})
			if err := json.Unmarshal(b, &g); err != nil {
				return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
			}
			if err := modifyGenesisTokenfactory(g, "tokenfactory", DenomMetadata_rupee, &roles, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisTokenfactory(g, "fiat-tokenfactory", DenomMetadata_drachma, &roles2, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisParamAuthority(g, paramauthorityWallet.Authority.Address); err != nil {
				return nil, err
			}
			out, err := json.Marshal(&g)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
			}
			return out, nil
		},
	}

	nv := 2
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			ChainConfig:   chainCfg,
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	noble = chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	kr := keyring.NewInMemory()

	fundedWallets := interchaintest.GetAndFundTestUsers(t, ctx, "broadcaster", 500000000000, noble)
	broadcasterWallet := fundedWallets[0]
	// number of wallets to create
	// this should ultimately but the amount of transcaction used the "load test"
	numWallets := 5

	wallets := make(map[int]ibc.Wallet, numWallets)

	var oneCoins types.Coins
	var coins types.Coins

	oneCoins = append(oneCoins, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.OneInt(),
	})

	coins = append(coins, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.NewInt(int64(numWallets)),
	})

	var inputs []sdkbanktypes.Input
	var outputs []sdkbanktypes.Output

	// Build, recover and prep wallets to be funded via a msgMultisend
	for i := 1; i <= numWallets; i++ {
		temp := interchaintest.BuildWallet(kr, fmt.Sprint(i), noble.Config())

		wallets[i] = ibc.Wallet{
			Mnemonic: temp.Mnemonic,
			Address:  temp.Bech32Address(chainCfg.Bech32Prefix),
			KeyName:  temp.KeyName,
			CoinType: temp.CoinType,
		}
		wallet := wallets[i]

		output := sdkbanktypes.Output{
			Address: wallet.Address,
			Coins:   oneCoins,
		}
		outputs = append(outputs, output)
	}

	input := sdkbanktypes.Output{
		Address: broadcasterWallet.Bech32Address(chainCfg.Bech32Prefix),
		Coins:   coins,
	}
	inputs = append(inputs, sdkbanktypes.Input(input))

	multisend := sdkbanktypes.MsgMultiSend{
		Inputs:  inputs,
		Outputs: outputs,
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		broadcasterWallet,
		&multisend,
	)
	require.NoError(t, err, "error broadcasting multisend tx")

	bal, err := noble.GetBalance(ctx, wallets[2].Address, chainCfg.Denom)
	require.NoError(t, err, "error getting balance")
	t.Log("BALANCE---1!!! ", bal)
	// All wallets are now funded

	toWallet := interchaintest.BuildWallet(kr, "toWallet", noble.Config())

	var eg errgroup.Group
	for _, wallet := range wallets {
		t.Log("FROM: ", wallet.Address)
		t.Log("FROM BECH ", wallet.Bech32Address(chainCfg.Bech32Prefix))
		t.Log("TO ", toWallet.Address)
		t.Log("TO BECH ", toWallet.Bech32Address(chainCfg.Bech32Prefix))

		wallet := wallet
		msgSend := sdkbanktypes.MsgSend{
			FromAddress: wallet.Address,
			ToAddress:   toWallet.Address,
			Amount:      oneCoins,
		}
		eg.Go(func() error {
			_, err = cosmos.BroadcastTx(
				ctx,
				broadcaster,
				&wallet,
				&msgSend,
			)
			if err != nil {
				return err
			}
			return nil
		})
	}
	require.NoError(t, eg.Wait())

	bal, err = noble.GetBalance(ctx, toWallet.Address, chainCfg.Denom)
	require.NoError(t, err, "failed to get balance")
	t.Log("BALANCE!!! ", bal)

	// emptyWallets := interchaintest.GetAndFundTestUsers(t, ctx, "new", 0, noble)
	// emptyWallet := emptyWallets[0]

	// send amount
	// send := ibc.WalletAmount{
	// 	Address: emptyWallet.Address,
	// 	Denom:   noble.Config().Denom,
	// 	Amount:  1,
	// }

	// duration := 5 * time.Second
	// startTime := time.Now()
	// endTime := startTime.Add(duration)

	// var counter int

	// for time.Now().Before(endTime) {
	// 	counter++
	// 	require.NoError(t, noble.SendFunds(ctx, extraWallets.User2.KeyName, send))
	// }

	// t.Log("Counter: ", counter)
	// testutil.WaitForBlocks(ctx, 5, noble)

	// threadedFunc(30, func() {
	// 	require.NoError(t, noble.SendFunds(ctx, extraWallets.User2.KeyName, send))
	// })

	// t.Log("BALANCE of new wallet after tx's!!!! ", bal)

	// bal, err = noble.GetBalance(ctx, string(address), noble.Config().Denom)
	// require.NoError(t, err, "error getting balance")
	// t.Log("Final Balance: ", bal)

}
