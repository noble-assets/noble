package ibctest_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/strangelove-ventures/ibctest/v3"
	"github.com/strangelove-ventures/ibctest/v3/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v3/ibc"
	"github.com/strangelove-ventures/ibctest/v3/relayer"
	"github.com/strangelove-ventures/ibctest/v3/relayer/rly"
	"github.com/strangelove-ventures/ibctest/v3/testreporter"
	"github.com/strangelove-ventures/ibctest/v3/testutil"
	integration "github.com/strangelove-ventures/noble/ibctest"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

type ForwardMetadata struct {
	Receiver       string        `json:"receiver"`
	Port           string        `json:"port"`
	Channel        string        `json:"channel"`
	Timeout        time.Duration `json:"timeout"`
	Retries        *uint8        `json:"retries,omitempty"`
	Next           *string       `json:"next,omitempty"`
	RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

func noblePreGenesis(ctx context.Context, val *cosmos.ChainNode) (string, error) {
	_, _, err := val.ExecBin(ctx, "add-consumer-section")
	if err != nil {
		return "", err
	}

	chainCfg := val.Chain.Config()

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

	err = val.RecoverKey(ctx, ownerKeyName, owner.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, masterMinterKeyName, masterMinter.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, minterControllerKeyName, minterController.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, minterKeyName, minter.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, blacklisterKeyName, blacklister.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, pauserKeyName, pauser.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, userKeyName, user.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, user2KeyName, user2.Mnemonic)
	if err != nil {
		return "", err
	}
	err = val.RecoverKey(ctx, aliceKeyName, alice.Mnemonic)
	if err != nil {
		return "", err
	}
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
		err = val.AddGenesisAccount(ctx, wallet.Address, []types.Coin{types.NewCoin(wallet.Denom, types.NewIntFromUint64(uint64(wallet.Amount)))})
		if err != nil {
			return "", err
		}
	}
	return owner.Address, nil
}

func TestPacketForwardMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	client, network := ibctest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	chainID_A, chainID_B, chainID_C, chainID_D := "chain-a", "chain-b", "chain-c", "chain-d"

	repo, version := integration.GetDockerImageInfo()

	var chainA, chainB, chainC, chainD ibc.Chain

	var ownerA, ownerB, ownerC, ownerD string

	nv := 1
	nf := 0
	coinType := "118"
	skipGenTx := true

	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*ibctest.ChainSpec{
		{
			NumValidators: &nv,
			NumFullNodes:  &nf,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "noble",
				ChainID:        chainID_A,
				Bin:            "nobled",
				Denom:          "token",
				Bech32Prefix:   "noble",
				CoinType:       coinType,
				SkipGenTx:      &skipGenTx,
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
					val := chainA.(*cosmos.CosmosChain).Validators[0]
					_, _, err := val.ExecBin(ctx, "add-consumer-section")
					if err != nil {
						return err
					}
					ownerA, err = noblePreGenesis(ctx, val)
					if err != nil {
						return err
					}
					return nil
				},
				ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
					return modifyGenesisNoble(b, ownerA)
				},
			}},
		{
			NumValidators: &nv,
			NumFullNodes:  &nf,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "noble",
				ChainID:        chainID_B,
				Bin:            "nobled",
				Denom:          "token",
				Bech32Prefix:   "noble",
				CoinType:       coinType,
				SkipGenTx:      &skipGenTx,
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
					val := chainB.(*cosmos.CosmosChain).Validators[0]
					_, _, err := val.ExecBin(ctx, "add-consumer-section")
					if err != nil {
						return err
					}
					ownerB, err = noblePreGenesis(ctx, val)
					if err != nil {
						return err
					}
					return nil
				},
				ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
					return modifyGenesisNoble(b, ownerB)
				},
			}},
		{
			NumValidators: &nv,
			NumFullNodes:  &nf,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "noble",
				ChainID:        chainID_C,
				Bin:            "nobled",
				Denom:          "token",
				Bech32Prefix:   "noble",
				CoinType:       coinType,
				SkipGenTx:      &skipGenTx,
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
					val := chainC.(*cosmos.CosmosChain).Validators[0]
					_, _, err := val.ExecBin(ctx, "add-consumer-section")
					if err != nil {
						return err
					}
					ownerC, err = noblePreGenesis(ctx, val)
					if err != nil {
						return err
					}
					return nil
				},
				ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
					return modifyGenesisNoble(b, ownerC)
				},
			}},
		{
			NumValidators: &nv,
			NumFullNodes:  &nf,
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "noble",
				ChainID:        chainID_D,
				Bin:            "nobled",
				Denom:          "token",
				Bech32Prefix:   "noble",
				CoinType:       coinType,
				SkipGenTx:      &skipGenTx,
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
					val := chainD.(*cosmos.CosmosChain).Validators[0]
					_, _, err := val.ExecBin(ctx, "add-consumer-section")
					if err != nil {
						return err
					}
					ownerD, err = noblePreGenesis(ctx, val)
					if err != nil {
						return err
					}
					return nil
				},
				ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
					return modifyGenesisNoble(b, ownerD)
				},
			}},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chainA, chainB, chainC, chainD = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain), chains[3].(*cosmos.CosmosChain)

	r := ibctest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		// TODO remove this line once default rly version includes https://github.com/cosmos/relayer/pull/1038
		relayer.CustomDockerImage("ghcr.io/cosmos/relayer", "andrew-trusting_period_ics", rly.RlyDefaultUidGid),
	).Build(t, client, network)

	const pathAB = "ab"
	const pathBC = "bc"
	const pathCD = "cd"

	ic := ibctest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddChain(chainC).
		AddChain(chainD).
		AddRelayer(r, "relayer").
		AddLink(ibctest.InterchainLink{
			Chain1:  chainA,
			Chain2:  chainB,
			Relayer: r,
			Path:    pathAB,
		}).
		AddLink(ibctest.InterchainLink{
			Chain1:  chainB,
			Chain2:  chainC,
			Relayer: r,
			Path:    pathBC,
		}).
		AddLink(ibctest.InterchainLink{
			Chain1:  chainC,
			Chain2:  chainD,
			Relayer: r,
			Path:    pathCD,
		})

	require.NoError(t, ic.Build(ctx, eRep, ibctest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	const userFunds = int64(10_000_000_000)
	users := ibctest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chainA, chainB, chainC, chainD)

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainID_A, chainID_B)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainID_C, chainID_B)
	require.NoError(t, err)

	bcChan := cbChan.Counterparty

	dcChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainID_D, chainID_C)
	require.NoError(t, err)

	cdChan := dcChan.Counterparty

	// Start the relayer on both paths
	err = r.StartRelayer(ctx, eRep, pathAB, pathBC, pathCD)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occured while stopping the relayer: %s", err)
			}
		},
	)

	// Get original account balances
	userA, userB, userC, userD := users[0], users[1], users[2], users[3]

	const transferAmount int64 = 100000

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, chainA.Config().Denom)
	secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)
	thirdHopDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, secondHopDenom)

	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)
	thirdHopDenomTrace := transfertypes.ParseDenomTrace(thirdHopDenom)

	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	secondHopIBCDenom := secondHopDenomTrace.IBCDenom()
	thirdHopIBCDenom := thirdHopDenomTrace.IBCDenom()

	firstHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainA.Config().Bech32Prefix, transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID))
	secondHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainB.Config().Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID))
	thirdHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainC.Config().Bech32Prefix, transfertypes.GetEscrowAddress(cdChan.PortID, abChan.ChannelID))

	t.Run("multi-hop a->b->c->d", func(t *testing.T) {
		// Send packet from Chain A->Chain B->Chain C->Chain D

		transfer := ibc.WalletAmount{
			Address: userB.Bech32Address(chainB.Config().Bech32Prefix),
			Denom:   chainA.Config().Denom,
			Amount:  transferAmount,
		}

		secondHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userD.Bech32Address(chainD.Config().Bech32Prefix),
				Channel:  cdChan.ChannelID,
				Port:     cdChan.PortID,
			},
		}
		nextBz, err := json.Marshal(secondHopMetadata)
		require.NoError(t, err)
		next := string(nextBz)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.Bech32Address(chainC.Config().Bech32Prefix),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Next:     &next,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName, transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+30, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainA)
		require.NoError(t, err)

		chainABalance, err := chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.Bech32Address(chainB.Config().Bech32Prefix), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.Bech32Address(chainC.Config().Bech32Prefix), secondHopIBCDenom)
		require.NoError(t, err)

		chainDBalance, err := chainD.GetBalance(ctx, userD.Bech32Address(chainD.Config().Bech32Prefix), thirdHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, userFunds-transferAmount, chainABalance)
		require.Equal(t, int64(0), chainBBalance)
		require.Equal(t, int64(0), chainCBalance)
		require.Equal(t, transferAmount, chainDBalance)

		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		thirdHopEscrowBalance, err := chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, transferAmount, firstHopEscrowBalance)
		require.Equal(t, transferAmount, secondHopEscrowBalance)
		require.Equal(t, transferAmount, thirdHopEscrowBalance)
	})

	t.Run("multi-hop denom unwind d->c->b->a", func(t *testing.T) {
		// Send packet back from Chain D->Chain C->Chain B->Chain A
		transfer := ibc.WalletAmount{
			Address: userC.Bech32Address(chainC.Config().Bech32Prefix),
			Denom:   thirdHopIBCDenom,
			Amount:  transferAmount,
		}

		secondHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userA.Bech32Address(chainA.Config().Bech32Prefix),
				Channel:  baChan.ChannelID,
				Port:     baChan.PortID,
			},
		}

		nextBz, err := json.Marshal(secondHopMetadata)
		require.NoError(t, err)

		next := string(nextBz)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userB.Bech32Address(chainB.Config().Bech32Prefix),
				Channel:  cbChan.ChannelID,
				Port:     cbChan.PortID,
				Next:     &next,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		require.NoError(t, err)

		chainDHeight, err := chainD.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainD.SendIBCTransfer(ctx, dcChan.ChannelID, userD.KeyName, transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainD, chainDHeight, chainDHeight+30, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainDBalance, err := chainD.GetBalance(ctx, userD.Bech32Address(chainD.Config().Bech32Prefix), thirdHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.Bech32Address(chainC.Config().Bech32Prefix), secondHopIBCDenom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.Bech32Address(chainB.Config().Bech32Prefix), firstHopIBCDenom)
		require.NoError(t, err)

		chainABalance, err := chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), chainA.Config().Denom)
		require.NoError(t, err)

		require.Equal(t, int64(0), chainDBalance)
		require.Equal(t, int64(0), chainCBalance)
		require.Equal(t, int64(0), chainBBalance)
		require.Equal(t, userFunds, chainABalance)

		// assert balances for IBC escrow accounts
		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		thirdHopEscrowBalance, err := chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, int64(0), firstHopEscrowBalance)
		require.Equal(t, int64(0), secondHopEscrowBalance)
		require.Equal(t, int64(0), thirdHopEscrowBalance)
	})

	t.Run("forward ack error refund", func(t *testing.T) {
		// Send a malformed packet with invalid receiver address from Chain A->Chain B->Chain C
		// This should succeed in the first hop and fail to make the second hop; funds should then be refunded to Chain A.
		transfer := ibc.WalletAmount{
			Address: userB.Bech32Address(chainB.Config().Bech32Prefix),
			Denom:   chainA.Config().Denom,
			Amount:  transferAmount,
		}

		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: "xyz1t8eh66t2w5k67kwurmn5gqhtq6d2ja0vp7jmmq", // malformed receiver address on Chain C
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName, transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.Bech32Address(chainB.Config().Bech32Prefix), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.Bech32Address(chainC.Config().Bech32Prefix), secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, userFunds, chainABalance)
		require.Equal(t, int64(0), chainBBalance)
		require.Equal(t, int64(0), chainCBalance)

		// assert balances for IBC escrow accounts
		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, int64(0), firstHopEscrowBalance)
		require.Equal(t, int64(0), secondHopEscrowBalance)
	})

	t.Run("forward timeout refund", func(t *testing.T) {
		// Send packet from Chain A->Chain B->Chain C with the timeout so low for B->C transfer that it can not make it from B to C, which should result in a refund from B to A after two retries.
		transfer := ibc.WalletAmount{
			Address: userB.Bech32Address(chainB.Config().Bech32Prefix),
			Denom:   chainA.Config().Denom,
			Amount:  transferAmount,
		}

		retries := uint8(2)
		metadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.Bech32Address(chainC.Config().Bech32Prefix),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Retries:  &retries,
				Timeout:  1 * time.Second,
			},
		}

		memo, err := json.Marshal(metadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName, transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+25, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainABalance, err := chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.Bech32Address(chainB.Config().Bech32Prefix), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.Bech32Address(chainC.Config().Bech32Prefix), secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, userFunds, chainABalance)
		require.Equal(t, int64(0), chainBBalance)
		require.Equal(t, int64(0), chainCBalance)

		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, int64(0), firstHopEscrowBalance)
		require.Equal(t, int64(0), secondHopEscrowBalance)
	})

	t.Run("multi-hop ack error refund", func(t *testing.T) {
		// Send a malformed packet with invalid receiver address from Chain A->Chain B->Chain C->Chain D
		// This should succeed in the first hop and second hop, then fail to make the third hop.
		// Funds should be refunded to Chain B and then to Chain A via acknowledgements with errors.
		transfer := ibc.WalletAmount{
			Address: userB.Bech32Address(chainB.Config().Bech32Prefix),
			Denom:   chainA.Config().Denom,
			Amount:  transferAmount,
		}

		secondHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: "xyz1t8eh66t2w5k67kwurmn5gqhtq6d2ja0vp7jmmq", // malformed receiver address on chain D
				Channel:  cdChan.ChannelID,
				Port:     cdChan.PortID,
			},
		}

		nextBz, err := json.Marshal(secondHopMetadata)
		require.NoError(t, err)

		next := string(nextBz)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.Bech32Address(chainC.Config().Bech32Prefix),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Next:     &next,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName, transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+30, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainDBalance, err := chainD.GetBalance(ctx, userD.Bech32Address(chainD.Config().Bech32Prefix), thirdHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.Bech32Address(chainC.Config().Bech32Prefix), secondHopIBCDenom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.Bech32Address(chainB.Config().Bech32Prefix), firstHopIBCDenom)
		require.NoError(t, err)

		chainABalance, err := chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), chainA.Config().Denom)
		require.NoError(t, err)

		require.Equal(t, userFunds, chainABalance)
		require.Equal(t, int64(0), chainBBalance)
		require.Equal(t, int64(0), chainCBalance)
		require.Equal(t, int64(0), chainDBalance)

		// assert balances for IBC escrow accounts
		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		thirdHopEscrowBalance, err := chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, int64(0), firstHopEscrowBalance)
		require.Equal(t, int64(0), secondHopEscrowBalance)
		require.Equal(t, int64(0), thirdHopEscrowBalance)
	})

	t.Run("multi-hop through native chain ack error refund", func(t *testing.T) {
		// send normal IBC transfer from B->A to get funds in IBC denom, then do multihop A->B(native)->C->D
		// this lets us test the burn from escrow account on chain C and the escrow to escrow transfer on chain B.

		// Compose the prefixed denoms and ibc denom for asserting balances
		baDenom := transfertypes.GetPrefixedDenom(abChan.PortID, abChan.ChannelID, chainB.Config().Denom)
		bcDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, chainB.Config().Denom)
		cdDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, bcDenom)

		baDenomTrace := transfertypes.ParseDenomTrace(baDenom)
		bcDenomTrace := transfertypes.ParseDenomTrace(bcDenom)
		cdDenomTrace := transfertypes.ParseDenomTrace(cdDenom)

		baIBCDenom := baDenomTrace.IBCDenom()
		bcIBCDenom := bcDenomTrace.IBCDenom()
		cdIBCDenom := cdDenomTrace.IBCDenom()

		transfer := ibc.WalletAmount{
			Address: userA.Bech32Address(chainA.Config().Bech32Prefix),
			Denom:   chainB.Config().Denom,
			Amount:  transferAmount,
		}

		chainBHeight, err := chainB.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainB.SendIBCTransfer(ctx, baChan.ChannelID, userB.KeyName, transfer, ibc.TransferOptions{})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainB, chainBHeight, chainBHeight+10, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainB)
		require.NoError(t, err)

		// assert balance for user controlled wallet
		chainABalance, err := chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), baIBCDenom)
		require.NoError(t, err)

		baEscrowBalance, err := chainB.GetBalance(
			ctx,
			sdk.MustBech32ifyAddressBytes(chainB.Config().Bech32Prefix, transfertypes.GetEscrowAddress(baChan.PortID, baChan.ChannelID)),
			chainB.Config().Denom,
		)
		require.NoError(t, err)

		require.Equal(t, transferAmount, chainABalance)
		require.Equal(t, transferAmount, baEscrowBalance)

		// Send a malformed packet with invalid receiver address from Chain A->Chain B->Chain C->Chain D
		// This should succeed in the first hop and second hop, then fail to make the third hop.
		// Funds should be refunded to Chain B and then to Chain A via acknowledgements with errors.
		transfer = ibc.WalletAmount{
			Address: userB.Bech32Address(chainB.Config().Bech32Prefix),
			Denom:   baIBCDenom,
			Amount:  transferAmount,
		}

		secondHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: "xyz1t8eh66t2w5k67kwurmn5gqhtq6d2ja0vp7jmmq", // malformed receiver address on chain D
				Channel:  cdChan.ChannelID,
				Port:     cdChan.PortID,
			},
		}

		nextBz, err := json.Marshal(secondHopMetadata)
		require.NoError(t, err)

		next := string(nextBz)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.Bech32Address(chainC.Config().Bech32Prefix),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Next:     &next,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err = chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName, transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+30, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 1, chainA)
		require.NoError(t, err)

		// assert balances for user controlled wallets
		chainDBalance, err := chainD.GetBalance(ctx, userD.Bech32Address(chainD.Config().Bech32Prefix), cdIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.Bech32Address(chainC.Config().Bech32Prefix), bcIBCDenom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.Bech32Address(chainB.Config().Bech32Prefix), chainB.Config().Denom)
		require.NoError(t, err)

		chainABalance, err = chainA.GetBalance(ctx, userA.Bech32Address(chainA.Config().Bech32Prefix), baIBCDenom)
		require.NoError(t, err)

		require.Equal(t, transferAmount, chainABalance)
		require.Equal(t, userFunds-transferAmount, chainBBalance)
		require.Equal(t, int64(0), chainCBalance)
		require.Equal(t, int64(0), chainDBalance)

		// assert balances for IBC escrow accounts
		cdEscrowBalance, err := chainC.GetBalance(
			ctx,
			sdk.MustBech32ifyAddressBytes(chainC.Config().Bech32Prefix, transfertypes.GetEscrowAddress(cdChan.PortID, cdChan.ChannelID)),
			bcIBCDenom,
		)
		require.NoError(t, err)

		bcEscrowBalance, err := chainB.GetBalance(
			ctx,
			sdk.MustBech32ifyAddressBytes(chainB.Config().Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID)),
			chainB.Config().Denom,
		)
		require.NoError(t, err)

		baEscrowBalance, err = chainB.GetBalance(
			ctx,
			sdk.MustBech32ifyAddressBytes(chainB.Config().Bech32Prefix, transfertypes.GetEscrowAddress(baChan.PortID, baChan.ChannelID)),
			chainB.Config().Denom,
		)
		require.NoError(t, err)

		require.Equal(t, transferAmount, baEscrowBalance)
		require.Equal(t, int64(0), bcEscrowBalance)
		require.Equal(t, int64(0), cdEscrowBalance)
	})
}
