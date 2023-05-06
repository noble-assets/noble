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

	// ---- PARAMETERS ---
	// Number of validators
	nv := 3
	// Number of full nodes
	nf := 3
	// number of wallets for each full node
	numWallets := 100
	// amount of times to loop through each wallet
	loop := 3
	// Channel buffer size per nf -- Amount of tx's to broadcast synchronously PER full node
	// So the total amount of synchronous broadcasts is (chnlBuff * nf)
	chnlBuff := 100
	// --- ---

	// total number of wallets
	totalWallets := numWallets * nf
	// total amount needed to fund all wallets - assumes we send 1 unit to each wallet during loop
	fundAmount := totalWallets * loop

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

	fundedWallets := interchaintest.GetAndFundTestUsers(t, ctx, "broadcaster", int64(fundAmount), noble)
	broadcasterWallet := fundedWallets[0]

	var totalAmountNeededToFundAllWallets types.Coins
	var amountToFundEachWallet types.Coins

	amountToFundEachWallet = append(amountToFundEachWallet, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.NewInt(int64(loop)),
	})

	totalAmountNeededToFundAllWallets = append(totalAmountNeededToFundAllWallets, types.Coin{
		Denom:  noble.Config().Denom,
		Amount: types.NewInt(int64(fundAmount)),
	})

	var inputs = make([]sdkbanktypes.Input, 0, 1)
	var outputs = make([]sdkbanktypes.Output, 0, totalWallets)

	var sliceOfOutputs = make([][]sdkbanktypes.Output, nf)

	var eg, eg2, eg3 errgroup.Group

	// creates wallets on each val and preps multiSend tx with addresses
	for n := 0; n < nf; n++ {
		n := n
		eg.Go(func() error {
			for i := 1; i <= numWallets; i++ {
				err := noble.FullNodes[n].CreateKey(ctx, fmt.Sprint(i))
				if err != nil {
					return err
				}
				add, err := noble.FullNodes[n].KeyBech32(ctx, fmt.Sprint(i), "")
				if err != nil {
					return err
				}
				output := sdkbanktypes.Output{
					Address: add,
					Coins:   amountToFundEachWallet,
				}
				sliceOfOutputs[n] = append(sliceOfOutputs[n], output)
			}
			return err
		})
	}
	require.NoError(t, eg.Wait())

	for i := 0; i < len(sliceOfOutputs); i++ {
		outputs = append(outputs, sliceOfOutputs[i]...)
	}

	input := sdkbanktypes.Input{
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

	// confirm tx went through by checking random wallet from each full node
	for n := 0; n < nf; n++ {
		rando := rand.Intn(numWallets)
		addr, err := noble.FullNodes[n].KeyBech32(ctx, fmt.Sprint(rando+1), "")
		require.NoError(t, err, "error getting key address")
		bal, err := noble.GetBalance(ctx, addr, chainCfg.Denom)
		require.NoError(t, err, "error getting balance")
		require.Equal(t, int64(loop), bal)
	}

	// All wallets are now funded

	kr := keyring.NewInMemory()

	// this is the wallet to send all transactions to during loop below
	toWallet := interchaintest.BuildWallet(kr, "toWallet", noble.Config())
	amount := fmt.Sprintf("1%s", noble.Config().Denom)

	startTimer := time.Now()

	sems := make([]chan struct{}, nf)
	for i := range sems {
		sems[i] = make(chan struct{}, chnlBuff) // create a buffered channel of size 2 for each nf
	}

	for n := 0; n < nf; n++ {
		n := n
		eg2.Go(func() error {
			for l := 1; l <= loop; l++ {
				t.Logf("\n\nSTATUS LOG ---> FN: %d - LOOP:%d/%d\n\n", n, l, loop)
				for i := 1; i <= numWallets; i++ {
					i := i
					sem := sems[n]
					sem <- struct{}{}
					eg3.Go(func() error {
						defer func() { <-sem }()
						_, _, err := noble.FullNodes[n].ExecBin(ctx, "tx", "bank", "send", fmt.Sprint(i), toWallet.Address, amount, "--keyring-backend", "test", "--node", fmt.Sprintf("tcp://%s:26657", noble.FullNodes[n].HostName()), "-b", "async", "-y")
						return err
					})
				}
			}
			return nil
		})
	}
	require.NoError(t, eg2.Wait())
	require.NoError(t, eg3.Wait())

	duration1 := time.Since(startTimer)

	// waits until all transactions have finalized
	bal, err := noble.GetBalance(ctx, toWallet.Address, chainCfg.Denom)
	for bal != int64(fundAmount) {
		t.Logf("Current Balance: %d Waiting for tx's to finalize...", bal)
		bal, err = noble.GetBalance(ctx, toWallet.Address, chainCfg.Denom)
		require.NoError(t, err, "failed to get balance")
		require.NoError(t, testutil.WaitForBlocks(ctx, 1, noble))
	}

	duration2 := time.Since(startTimer)

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

	t.Log("\n\n --------- \n\n")
	t.Logf("%d TRANSACTIONS BROADCASTED IN %v", bal, duration1)
	t.Logf("TRANSACTIONS BROADCASTED AND FINALIZED IN %v", duration2)
	t.Logf("AVG TRANSACTIONS PER SECOND: %f", float64(bal)/float64(duration1.Seconds()))
	t.Logf("AVG TRANSACTIONS PER BLOCK %f", averageTxPerBlock)
	t.Logf("AVG BLOCKTIME: %f", avgBlockTime)

}
