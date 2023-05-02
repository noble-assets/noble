package interchaintest_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdkbanktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/interchaintest/v3/testutil"

	"github.com/strangelove-ventures/noble/cmd"
	integration "github.com/strangelove-ventures/noble/interchaintest"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	_ "modernc.org/sqlite"
)

func TestLoad(t *testing.T) {
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

	configTomlOverrides := make(testutil.Toml)

	consensus := make(testutil.Toml)
	// blockTime := reset to chain defaults (not interchain test defaults)
	consensus["timeout_commit"] = (time.Duration(5) * time.Second).String()
	consensus["timeout_propose"] = (time.Duration(3) * time.Second).String()

	configTomlOverrides["consensus"] = consensus

	rpc := make(testutil.Toml)
	rpc["max_subscription_clients"] = "500"

	configTomlOverrides["rpc"] = rpc

	configFileOverrides := make(map[string]any)
	configFileOverrides["config/config.toml"] = configTomlOverrides

	chainCfg := ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "noble",
		ChainID:             "noble-1",
		Bin:                 "nobled",
		Denom:               "utoken",
		Bech32Prefix:        "noble",
		CoinType:            "118",
		GasPrices:           "0.0token",
		GasAdjustment:       1.1,
		TrustingPeriod:      "504h",
		NoHostMount:         false,
		ConfigFileOverrides: configFileOverrides,
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

	nv := 2 // Number of validators
	nf := 0 // Number of full nodes

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

	cwd, err := os.Getwd()
	require.NoError(t, err, "error getting cwd")

	dbFileName := "block.db"
	dbFolder := filepath.Join(cwd, "ictest_db")
	dbFileFullPath := filepath.Join(dbFolder, dbFileName)
	require.NoError(t, os.RemoveAll(dbFolder))

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: dbFileFullPath,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	kr := keyring.NewInMemory()

	// ---- PARAMETERS ---

	// number of wallets to create
	numWallets := 500
	// amount of times to loop through each wallet
	loop := 5
	// total amount needed to fund all wallets
	// assumes we send 1 unit to each wallet during loop
	fundAmount := numWallets * loop

	// --- ---

	fundedWallets := interchaintest.GetAndFundTestUsers(t, ctx, "broadcaster", int64(fundAmount), noble)
	broadcasterWallet := fundedWallets[0]

	wallets := make(map[int]ibc.Wallet, numWallets)

	var totalAmountNeededToFundAllWallets types.Coins
	var amountToFundEachWallet types.Coins
	var oneCoins types.Coins

	oneCoins = append(oneCoins, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.OneInt(),
	})

	amountToFundEachWallet = append(amountToFundEachWallet, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.NewInt(int64(loop)),
	})

	totalAmountNeededToFundAllWallets = append(totalAmountNeededToFundAllWallets, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.NewInt(int64(fundAmount)),
	})

	var inputs []sdkbanktypes.Input
	var outputs []sdkbanktypes.Output

	// Build, recover and prep wallets to be funded via a msgMultisend
	for i := 1; i <= numWallets; i++ {
		wallets[i] = interchaintest.BuildWallet(kr, fmt.Sprint(i), noble.Config())
		require.NoError(t, noble.RecoverKey(ctx, fmt.Sprint(i), wallets[i].Mnemonic), "failed to recover key")

		output := sdkbanktypes.Output{
			Address: wallets[i].Address,
			Coins:   amountToFundEachWallet,
		}
		outputs = append(outputs, output)
	}

	input := sdkbanktypes.Output{
		Address: broadcasterWallet.Bech32Address(chainCfg.Bech32Prefix),
		Coins:   totalAmountNeededToFundAllWallets,
	}
	inputs = append(inputs, sdkbanktypes.Input(input))

	multisend := sdkbanktypes.MsgMultiSend{
		Inputs:  inputs,
		Outputs: outputs,
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)
	f, err := broadcaster.GetFactory(ctx, broadcasterWallet)
	require.NoError(t, err, "error getting broadcaster factory")
	broadcaster.ConfigureFactoryOptions(func(factory tx.Factory) tx.Factory {
		return f.WithSimulateAndExecute(true)
	})

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		broadcasterWallet,
		&multisend,
	)
	require.NoError(t, err, "error broadcasting multisend tx")
	// confirm tx went through by checking random wallet
	bal, err := noble.GetBalance(ctx, wallets[rand.Intn(len(wallets))].Address, chainCfg.Denom)
	require.NoError(t, err, "error getting account balance")
	require.Equal(t, int64(loop), bal)

	// All wallets are now funded

	// this is the wallet to send all transactions to during loop below
	toWallet := interchaintest.BuildWallet(kr, "toWallet", noble.Config())

	var eg errgroup.Group

	broadcaster.ConfigureFactoryOptions(func(factory tx.Factory) tx.Factory {
		return f.WithSimulateAndExecute(false)
	})

	startTimer := time.Now()

	for i := 1; i <= loop; i++ {
		t.Logf("starting loop: %d...", i)
		for _, wallet := range wallets {
			wallet := wallet

			eg.Go(func() error {
				msgSend := sdkbanktypes.MsgSend{
					FromAddress: wallet.Address,
					ToAddress:   toWallet.Address,
					Amount:      oneCoins,
				}

				decoded := types.MustAccAddressFromBech32(wallet.Address)
				wallet.Address = string(decoded)

				_, err := cosmos.BroadcastTx(
					ctx,
					cosmos.NewBroadcaster(t, noble),
					&wallet,
					&msgSend,
				)
				return err
			})
		}
		require.NoError(t, eg.Wait())
	}

	duration := time.Since(startTimer)

	// testutil.WaitForBlocks(ctx, 2, noble)

	bal, err = noble.GetBalance(ctx, toWallet.Address, chainCfg.Denom)
	require.NoError(t, err, "failed to get balance")

	db, err := sql.Open("sqlite", dbFileFullPath)
	require.NoError(t, err, "error opening sql db")

	var sqlMsg CosmosMessageResult

	row := db.QueryRow("SELECT block_height FROM v_cosmos_messages where type=?", "/cosmos.bank.v1beta1.MsgMultiSend")
	require.NoError(t, row.Scan(&sqlMsg.Height), "failed to get height")

	// after the multi send, we start rapid firing tx's
	// for the accuracy of the averages computed below, lets only consider
	// blocks where we are broadcasting for the full block time.
	heightRangeStart := sqlMsg.Height + 2

	row = db.QueryRow("SELECT block_height FROM v_cosmos_messages ORDER BY block_height DESC")
	require.NoError(t, row.Scan(&sqlMsg.Height), "failed to get height")

	// for the accuracy of the averages computed below, lets only consider
	// blocks where we are broadcasting for the full block time.
	heightRangeEnd := sqlMsg.Height - 1

	var count int
	var txsPerBlock []int
	for i := heightRangeStart; i <= heightRangeEnd; i++ {
		row = db.QueryRow("SELECT COUNT(block_height) FROM v_cosmos_messages WHERE block_height = ?", i)
		require.NoErrorf(t, row.Scan(&count), "error counting on block %d", i)
		txsPerBlock = append(txsPerBlock, count)
	}

	var sum int
	for i := 0; i < len(txsPerBlock); i++ {
		sum += txsPerBlock[i]
	}
	averageTxPerBlock := float32(sum) / float32(len(txsPerBlock))

	rows, err := db.Query("SELECT created_at FROM block")
	require.NoError(t, err, "failed to get block times from sql db")

	var createdAtTimes []time.Time
	for rows.Next() {
		var createdAt string
		require.NoError(t, rows.Scan(&createdAt), "error querying created_at from sql")
		timeParse, err := time.Parse(time.RFC3339, createdAt)
		require.NoError(t, err, "error parsing time string")
		createdAtTimes = append(createdAtTimes, timeParse)
	}
	var blocktimes []float64
	for i := heightRangeStart; i < heightRangeEnd-1; i++ {
		timeSub := createdAtTimes[i+1].Sub(createdAtTimes[i])
		blocktimes = append(blocktimes, timeSub.Seconds())
	}
	var sumBlockTimes float64
	for i := 0; i < len(blocktimes); i++ {
		sumBlockTimes += blocktimes[i]
	}
	avgBlockTime := float32(sumBlockTimes) / float32(len(blocktimes))

	t.Logf("%d TRANSACTIONS BROADCASTED IN %v", bal, duration)
	t.Logf("AVG TRANSACTIONS PER SECOND: %f", float64(bal)/float64(duration.Seconds()))
	t.Logf("AVG TRANSACTIONS PER BLOCK %f", averageTxPerBlock)
	t.Logf("AVG BLOCKTIME: %f", avgBlockTime)

}

type CosmosMessageResult struct {
	Height int64
	Index  int
	Type   string // URI for proto definition, e.g. /ibc.core.client.v1.MsgCreateClient

	ClientChainID sql.NullString

	ClientID             sql.NullString
	CounterpartyClientID sql.NullString

	ConnID             sql.NullString
	CounterpartyConnID sql.NullString

	PortID             sql.NullString
	CounterpartyPortID sql.NullString

	ChannelID             sql.NullString
	CounterpartyChannelID sql.NullString
}
