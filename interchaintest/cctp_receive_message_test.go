package interchaintest_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"sort"
	"testing"
	"time"

	cosmossdk_io_math "cosmossdk.io/math"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/interchaintest/v3/testutil"
	routertypes "github.com/strangelove-ventures/noble-router/x/router/types"
	"github.com/strangelove-ventures/noble/cmd"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestCCTP_ReceiveMessage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var gw genesisWrapper

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", nv, nf, true, false, true, false),
		{
			Name:          "gaia",
			Version:       "v10.0.2",
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain
	gaia := chains[1].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayerImage,
	).Build(t, client, network)

	pathName := "noble-gaia"

	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(r, "r").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: r,
			Path:    pathName,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
		// BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	nobleChainCfg := noble.Config()
	gaiaChainCfg := gaia.Config()

	cmd.SetPrefixes(nobleChainCfg.Bech32Prefix)

	attesters := make([]*ecdsa.PrivateKey, 2)
	msgs := make([]sdk.Msg, 2)

	// attester - ECDSA public key (Circle will own these keys for mainnet)
	for i := range attesters {
		p, err := crypto.GenerateKey() // private key
		require.NoError(t, err)

		attesters[i] = p

		pubKey := elliptic.Marshal(p.PublicKey, p.PublicKey.X, p.PublicKey.Y) //public key

		attesterPub := hex.EncodeToString(pubKey)

		// Adding an attester to protocal
		msgs[i] = &cctptypes.MsgEnableAttester{
			From:     gw.fiatTfRoles.Owner.FormattedAddress(),
			Attester: attesterPub,
		}
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)
	broadcaster.ConfigureClientContextOptions(func(clientContext sdkclient.Context) sdkclient.Context {
		return clientContext.WithBroadcastMode(flags.BroadcastBlock)
	})

	t.Log("preparing to submit add public keys tx")

	burnToken := make([]byte, 32)
	copy(burnToken[12:], common.FromHex("0x07865c6E87B9F70255377e024ace6630C1Eaa37F"))

	// maps remote token on remote domain to a local token -- used for minting
	msgs = append(msgs, &cctptypes.MsgLinkTokenPair{
		From:         gw.fiatTfRoles.Owner.FormattedAddress(),
		RemoteDomain: 0,
		RemoteToken:  burnToken,
		LocalToken:   denomMetadataDrachma.Base,
	})

	bCtx, bCancel := context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err := cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		msgs...,
	)
	require.NoError(t, err, "error submitting add public keys tx")
	require.Zero(t, tx.Code, "cctp add pub keys transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("Submitted add public keys tx: %s", tx.TxHash)

	var senderForward = []byte("12345678901234567890123456789015")

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		gw.paramAuthority,
		&routertypes.MsgAddAllowedSourceDomainSender{
			From:     gw.paramAuthority.FormattedAddress(),
			DomainId: 0,
			Address:  senderForward,
		},
	)
	require.NoError(t, err, "error submitting add allowed source domain sender tx")
	require.Zero(t, tx.Code, "cctp add allowed source domain sender transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	nobleValidator := noble.Validators[0]

	cctpModuleAccount := authtypes.NewModuleAddress(cctptypes.ModuleName).String()

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", gw.fiatTfRoles.MinterController.FormattedAddress(), cctpModuleAccount, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", cctpModuleAccount, "1000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	const receiver = "9B6CA0C13EB603EF207C4657E1E619EF531A4D27" //account

	receiverBz, err := hex.DecodeString(receiver)
	require.NoError(t, err)

	nobleReceiver, err := bech32.ConvertAndEncode(nobleChainCfg.Bech32Prefix, receiverBz)
	require.NoError(t, err)

	gaiaReceiver, err := bech32.ConvertAndEncode(gaiaChainCfg.Bech32Prefix, receiverBz)
	require.NoError(t, err)

	burnRecipientPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, receiverBz...)

	// someone burned USDC on Ethereum -> Mint on Noble
	depositForBurn := cctptypes.BurnMessage{
		BurnToken:     burnToken,
		MintRecipient: burnRecipientPadded,
		Amount:        cosmossdk_io_math.NewInt(1000000),
		MessageSender: burnRecipientPadded,
	}

	depositForBurnBz, err := depositForBurn.Bytes()
	require.NoError(t, err)

	var senderBurn = []byte("12345678901234567890123456789012")

	emptyDestinationCaller := make([]byte, 32)

	wrappedDepositForBurn := cctptypes.Message{
		Version:           0,
		SourceDomain:      0,
		DestinationDomain: 4, // Noble is 4
		Nonce:             0, // dif per message
		Sender:            senderBurn,
		Recipient:         cctptypes.PaddedModuleAddress,
		DestinationCaller: emptyDestinationCaller,
		MessageBody:       depositForBurnBz,
	}

	wrappedDepositForBurnBz, err := wrappedDepositForBurn.Bytes()
	require.NoError(t, err)

	// in mainnet this would forward to dydx chain
	forward := routertypes.IBCForwardMetadata{
		Port:                "transfer",
		Channel:             "channel-0",
		DestinationReceiver: gaiaReceiver,
	}

	forwardBz, err := forward.Bytes(gaiaChainCfg.Bech32Prefix)
	require.NoError(t, err)

	wrappedForward := cctptypes.Message{
		Version:           0,
		SourceDomain:      0, // same source domain !
		DestinationDomain: 4,
		Nonce:             1,             // cant be same nonce as above
		Sender:            senderForward, // different sender !
		Recipient:         burnRecipientPadded,
		DestinationCaller: emptyDestinationCaller,
		MessageBody:       forwardBz,
	}

	wrappedForwardBz, err := wrappedForward.Bytes()
	require.NoError(t, err)

	digestBurn := crypto.Keccak256(wrappedDepositForBurnBz) // hashed message is the key to the attestation
	digestForward := crypto.Keccak256(wrappedForwardBz)

	attestationBurn := make([]byte, 0, len(attesters)*65) //65 byte
	attestationForward := make([]byte, 0, len(attesters)*65)

	// CCTP requires attestations to have signatures sorted by address
	sort.Slice(attesters, func(i, j int) bool {
		return bytes.Compare(
			crypto.PubkeyToAddress(attesters[i].PublicKey).Bytes(),
			crypto.PubkeyToAddress(attesters[j].PublicKey).Bytes(),
		) < 0
	})

	for i := range attesters {
		sig, err := crypto.Sign(digestBurn, attesters[i])
		require.NoError(t, err)

		attestationBurn = append(attestationBurn, sig...)

		sig, err = crypto.Sign(digestForward, attesters[i])
		require.NoError(t, err)

		attestationForward = append(attestationForward, sig...)
	}

	t.Logf("Attested to messages: %s", tx.TxHash)

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()
	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		&cctptypes.MsgReceiveMessage{ //note: all messages that go to noble go through MsgReceiveMessage
			From:        gw.fiatTfRoles.Owner.FormattedAddress(),
			Message:     wrappedDepositForBurnBz,
			Attestation: attestationBurn,
		},
	)
	require.NoError(t, err, "error submitting cctp burn recv tx")
	require.Zerof(t, tx.Code, "cctp burn recv transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP burn message successfully received: %s", tx.TxHash)

	balance, err := noble.GetBalance(ctx, nobleReceiver, denomMetadataDrachma.Base)
	require.NoError(t, err)

	require.Equal(t, int64(1000000), balance)

	err = r.StartRelayer(ctx, eRep, pathName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = r.StopRelayer(ctx, eRep)
	})

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()
	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		&cctptypes.MsgReceiveMessage{
			From:        gw.fiatTfRoles.Owner.FormattedAddress(),
			Message:     wrappedForwardBz,
			Attestation: attestationForward,
		},
	)

	require.NoError(t, err, "error submitting cctp forward recv tx")
	require.Zerof(t, tx.Code, "cctp forward recv transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP IBC forward message successfully received: %s", tx.TxHash)

	err = testutil.WaitForBlocks(ctx, 10, noble, gaia)
	require.NoError(t, err)

	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", denomMetadataDrachma.Base))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	gaiaBal, err := gaia.GetBalance(ctx, gaiaReceiver, dstIbcDenom)
	require.NoError(t, err)
	require.Equal(t, int64(999900), gaiaBal)

	// now test deposit for burn noble -> eth

	// gaia.Validators[0].CreateKey(ctx, "my_key")

	// send ibc tx back to noble
	// gaia.Validators[0].SendIBCTransfer(ctx, "channel-0", )

	// depositForBurnNoble := &cctptypes.MsgDepositForBurn{
	// 	BurnToken:uusdc
	// 	Amount:    cosmossdk_io_math.NewInt(1000000),
	// }

	// broad cast... take tx..
	// tx.event

	// make sure balance went down
}
