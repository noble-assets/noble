package interchaintest_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"sort"
	"testing"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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

	"github.com/strangelove-ventures/noble/x/cctp/keeper"
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
		nobleChainSpec(ctx, &gw, "noble-1", nv, nf, true, true, true, true),
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

	msgs = append(msgs, &cctptypes.MsgLinkTokenPair{
		From:         gw.fiatTfRoles.Owner.FormattedAddress(),
		RemoteDomain: 0,
		RemoteToken:  "07865c6E87B9F70255377e024ace6630C1Eaa37F",
		LocalToken:   "uusdc",
	})

	tx, err := cosmos.BroadcastTx(
		ctx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		msgs...,
	)
	require.NoError(t, err, "error submitting add public keys tx")
	require.Zero(t, tx.Code)

	const receiver = "9B6CA0C13EB603EF207C4657E1E619EF531A4D27"

	receiverBz, err := hex.DecodeString(receiver)
	require.NoError(t, err)

	nobleReceiver, err := bech32.ConvertAndEncode(nobleChainCfg.Bech32Prefix, receiverBz)
	require.NoError(t, err)

	gaiaReceiver, err := bech32.ConvertAndEncode(gaiaChainCfg.Bech32Prefix, receiverBz)
	require.NoError(t, err)

	burnRecipient := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, receiverBz...)

	depositForBurn := keeper.ParseBurnMessageIntoBytes(cctptypes.BurnMessage{
		BurnToken:     []byte("07865c6E87B9F70255377e024ace6630C1Eaa37F"),
		MintRecipient: burnRecipient,
		Amount:        1000000,
		MessageSender: receiverBz,
	})

	forward, err := proto.Marshal(&routertypes.IBCForwardMetadata{
		Port:                "transfer",
		Channel:             "channel-0",
		DestinationReceiver: gaiaReceiver,
	})
	require.NoError(t, err)

	// depositForBurn, err := hex.DecodeString("0000000000000000000000010000000000039148000000000000000000000000D0C3DA58F55358142B8D3E06C1C30C5C6114EFE8000000000000000000000000EB08F243E5D3FCFF26A9E38AE5520A669F4019D000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007865C6E87B9F70255377E024ACE6630C1EAA37F0000000000000000000000009B6CA0C13EB603EF207C4657E1E619EF531A4D2700000000000000000000000000000000000000000000000000000000000F42400000000000000000000000009B6CA0C13EB603EF207C4657E1E619EF531A4D27")
	// require.NoError(t, err)

	digest := crypto.Keccak256(depositForBurn)

	attestation := make([]byte, 0, len(attesters)*65)

	// CCTP requires attestations to have signatures sorted by address
	sort.Slice(attesters, func(i, j int) bool {
		return bytes.Compare(
			crypto.PubkeyToAddress(attesters[i].PublicKey).Bytes(),
			crypto.PubkeyToAddress(attesters[j].PublicKey).Bytes(),
		) < 0
	})

	for i := range attesters {
		sig, err := crypto.Sign(digest, attesters[i])
		require.NoError(t, err)

		attestation = append(attestation, sig...)
	}

	tx, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		&cctptypes.MsgReceiveMessage{
			From:        gw.fiatTfRoles.Owner.FormattedAddress(),
			Message:     depositForBurn,
			Attestation: attestation,
		},
	)
	require.NoError(t, err, "error submitting cctp recv tx")
	require.Zerof(t, tx.Code, "cctp recv transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP message successfully received: %s", tx.TxHash)

	balance, err := noble.GetBalance(ctx, nobleReceiver, "uusdc")
	require.NoError(t, err)

	require.Equal(t, int64(1000000), balance)

	err = testutil.WaitForBlocks(ctx, 100, noble)
	require.NoError(t, err)
}
