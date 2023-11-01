package interchaintest_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"sort"
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/crypto"

	cosmossdk_io_math "cosmossdk.io/math"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/noble/v4/cmd"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestCCTP_ReplaceDepositForBurn(t *testing.T) {
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

	ic := interchaintest.NewInterchain().
		AddChain(noble)

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

	burnRecipientPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, receiverBz...)

	messageSender := make([]byte, 32)
	copy(messageSender[12:], sdk.MustAccAddressFromBech32(gw.extraWallets.User.FormattedAddress()))

	// someone burned USDC on Ethereum -> Mint on Noble
	depositForBurn := cctptypes.BurnMessage{
		BurnToken:     burnToken,
		MintRecipient: burnRecipientPadded,
		Amount:        cosmossdk_io_math.NewInt(1000000),
		MessageSender: messageSender,
	}

	depositForBurnBz, err := depositForBurn.Bytes()
	require.NoError(t, err)

	emptyDestinationCaller := make([]byte, 32)

	wrappedDepositForBurn := cctptypes.Message{
		Version:           0,
		SourceDomain:      4, // noble is 4
		DestinationDomain: 0,
		Nonce:             0, // dif per message
		Sender:            cctptypes.PaddedModuleAddress,
		Recipient:         cctptypes.PaddedModuleAddress,
		DestinationCaller: emptyDestinationCaller,
		MessageBody:       depositForBurnBz,
	}

	wrappedDepositForBurnBz, err := wrappedDepositForBurn.Bytes()
	require.NoError(t, err)

	digestBurn := crypto.Keccak256(wrappedDepositForBurnBz) // hashed message is the key to the attestation

	attestationBurn := make([]byte, 0, len(attesters)*65) //65 byte

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
	}

	t.Logf("Attested to messages: %s", tx.TxHash)

	newDestCaller := []byte("12345678901234567890123456789012")
	newMintRecipient := []byte("12345678901234567890123456789012")

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()
	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		gw.extraWallets.User,
		&cctptypes.MsgReplaceDepositForBurn{
			From:                 gw.extraWallets.User.FormattedAddress(),
			OriginalMessage:      wrappedDepositForBurnBz,
			OriginalAttestation:  attestationBurn,
			NewDestinationCaller: newDestCaller,
			NewMintRecipient:     newMintRecipient,
		},
	)
	require.NoError(t, err, "error submitting cctp replace deposit for burn tx")
	require.Zerof(t, tx.Code, "cctp replace deposit for burn transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP replace message successfully received: %s", tx.TxHash)

	for _, rawEvent := range tx.Events {
		switch rawEvent.Type {
		case "circle.cctp.v1.DepositForBurn":
			parsedEvent, err := sdk.ParseTypedEvent(rawEvent)
			require.NoError(t, err)
			actualDepositForBurn, ok := parsedEvent.(*cctptypes.DepositForBurn)
			require.True(t, ok)

			expectedBurnToken := hex.EncodeToString(crypto.Keccak256(depositForBurn.BurnToken))

			require.Equal(t, wrappedDepositForBurn.Nonce, actualDepositForBurn.Nonce)
			require.Equal(t, expectedBurnToken, actualDepositForBurn.BurnToken)
			require.Equal(t, depositForBurn.Amount, actualDepositForBurn.Amount)
			require.Equal(t, gw.extraWallets.User.FormattedAddress(), actualDepositForBurn.Depositor)
			require.Equal(t, newMintRecipient, actualDepositForBurn.MintRecipient) // new
			require.Equal(t, wrappedDepositForBurn.DestinationDomain, actualDepositForBurn.DestinationDomain)
			require.Equal(t, wrappedDepositForBurn.Recipient, actualDepositForBurn.DestinationTokenMessenger)
			require.Equal(t, newDestCaller, actualDepositForBurn.DestinationCaller) // new
		case "circle.cctp.v1.MessageSent":
			parsedEvent, err := sdk.ParseTypedEvent(rawEvent)
			require.NoError(t, err)
			event, ok := parsedEvent.(*cctptypes.MessageSent)
			require.True(t, ok)

			message, err := new(cctptypes.Message).Parse(event.Message)
			require.NoError(t, err)

			expectedBurnToken := hex.EncodeToString(crypto.Keccak256(depositForBurn.BurnToken))
			fmt.Println(expectedBurnToken)

			moduleAddress := make([]byte, 32)
			copy(moduleAddress[12:], sdk.MustAccAddressFromBech32(gw.extraWallets.User.FormattedAddress()))

			require.Equal(t, wrappedDepositForBurn.Version, message.Version)
			require.Equal(t, wrappedDepositForBurn.SourceDomain, message.SourceDomain)
			require.Equal(t, wrappedDepositForBurn.DestinationDomain, message.DestinationDomain)
			require.Equal(t, wrappedDepositForBurn.Nonce, message.Nonce)
			require.Equal(t, cctptypes.PaddedModuleAddress, message.Sender)
			require.Equal(t, cctptypes.PaddedModuleAddress, message.Recipient)
			require.Equal(t, newDestCaller, message.DestinationCaller)

			body, err := new(cctptypes.BurnMessage).Parse(message.MessageBody)
			require.NoError(t, err)

			require.Equal(t, depositForBurn.Version, body.Version)
			require.Equal(t, newMintRecipient, body.MintRecipient)
			require.Equal(t, depositForBurn.Amount, body.Amount)
			require.True(t, bytes.Equal(depositForBurn.BurnToken, body.BurnToken))
			require.Equal(t, depositForBurn.MessageSender, body.MessageSender)
		}
	}
}
