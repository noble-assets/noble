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

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/interchaintest/v3/testutil"
	"github.com/strangelove-ventures/noble/cmd"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/strangelove-ventures/noble/x/cctp/types"
	cctptypes "github.com/strangelove-ventures/noble/x/cctp/types"
	routertypes "github.com/strangelove-ventures/noble/x/router/types"
)

// run `make local-image`to rebuild updated binary before running test
func TestCCTP(t *testing.T) {
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

		// TODO set to false
		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	nobleChainCfg := noble.Config()
	gaiaChainCfg := gaia.Config()

	cmd.SetPrefixes(nobleChainCfg.Bech32Prefix)

	attesters := make([]*ecdsa.PrivateKey, 2)
	msgs := make([]sdk.Msg, 2)

	for i := range attesters {
		p, err := crypto.GenerateKey()
		require.NoError(t, err)

		attesters[i] = p

		pubKey := elliptic.Marshal(p.PublicKey, p.PublicKey.X, p.PublicKey.Y)

		attesterPub := hex.EncodeToString(pubKey)

		msgs[i] = &cctptypes.MsgAddPublicKey{
			From:      gw.fiatTfRoles.Owner.FormattedAddress(),
			PublicKey: []byte(attesterPub),
		}
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)
	broadcaster.ConfigureClientContextOptions(func(clientContext sdkclient.Context) sdkclient.Context {
		return clientContext.WithBroadcastMode(flags.BroadcastBlock)
	})

	t.Log("preparing to submit add public keys tx")

	msgs = append(msgs, &cctptypes.MsgLinkTokenPair{
		From:         gw.fiatTfRoles.Owner.FormattedAddress(),
		RemoteDomain: 0,
		RemoteToken:  "07865c6E87B9F70255377e024ace6630C1Eaa37F",
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

	nobleValidator := noble.Validators[0]

	cctpModuleAccount := authtypes.NewModuleAddress(types.ModuleName).String()

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", gw.fiatTfRoles.MinterController.FormattedAddress(), cctpModuleAccount, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", cctpModuleAccount, "1000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	const receiver = "9B6CA0C13EB603EF207C4657E1E619EF531A4D27"

	receiverBz, err := hex.DecodeString(receiver)
	require.NoError(t, err)

	nobleReceiver, err := bech32.ConvertAndEncode(nobleChainCfg.Bech32Prefix, receiverBz)
	require.NoError(t, err)

	gaiaReceiver, err := bech32.ConvertAndEncode(gaiaChainCfg.Bech32Prefix, receiverBz)
	require.NoError(t, err)

	burnRecipientPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, receiverBz...)

	burnToken, err := hex.DecodeString("07865c6E87B9F70255377e024ace6630C1Eaa37F")
	require.NoError(t, err)

	burnTokenPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, burnToken...)

	depositForBurn := &cctptypes.BurnMessage{
		BurnToken:     burnTokenPadded,
		MintRecipient: burnRecipientPadded,
		Amount:        1000000,
		MessageSender: receiverBz,
	}
	depositForBurnBz := depositForBurn.Bytes()

	forward, err := proto.Marshal(&routertypes.IBCForwardMetadata{
		Port:                "transfer",
		Channel:             "channel-0",
		DestinationReceiver: gaiaReceiver,
	})
	require.NoError(t, err)

	wrappedDepositForBurn := cctptypes.Message{
		Version:           0,
		SourceDomain:      0,
		DestinationDomain: 4,
		Nonce:             0,
		Sender:            []byte("12345678901234567890123456789012"),
		Recipient:         []byte(nobleReceiver),
		// DestinationCaller: []byte(gw.fiatTfRoles.Owner.FormattedAddress()),
		MessageBody: depositForBurnBz,
	}
	wrappedDepositForBurnBz := wrappedDepositForBurn.Bytes()

	wrappedForward := &cctptypes.Message{
		Version:           0,
		SourceDomain:      0,
		DestinationDomain: 4,
		Nonce:             1,
		Sender:            []byte("12345678901234567890123456789012"),
		Recipient:         []byte(nobleReceiver),
		// DestinationCaller: []byte(gw.fiatTfRoles.Owner.FormattedAddress()),
		MessageBody: forward,
	}
	wrappedForwardBz := wrappedForward.Bytes()

	digestBurn := crypto.Keccak256(wrappedDepositForBurnBz)
	digestForward := crypto.Keccak256(wrappedForwardBz)

	attestationBurn := make([]byte, 0, len(attesters)*65)
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

	tx, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		&cctptypes.MsgReceiveMessage{
			From:        gw.fiatTfRoles.Owner.FormattedAddress(),
			Message:     wrappedDepositForBurnBz,
			Attestation: attestationBurn,
		},
	)
	require.NoError(t, err, "error submitting cctp burn recv tx")
	require.Zerof(t, tx.Code, "cctp burn recv transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP message successfully received: %s", tx.TxHash)

	balance, err := noble.GetBalance(ctx, nobleReceiver, denomMetadataDrachma.Base)
	require.NoError(t, err)

	require.Equal(t, int64(1000000), balance)

	tx, err = cosmos.BroadcastTx(
		ctx,
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

	err = testutil.WaitForBlocks(ctx, 100, noble)
	require.NoError(t, err)
}
