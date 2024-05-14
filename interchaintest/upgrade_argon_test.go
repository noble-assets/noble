package interchaintest_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
)

func testPostArgonUpgrade(
	t *testing.T,
	ctx context.Context,
	noble *cosmos.CosmosChain,
	paramAuthority ibc.Wallet,
) {
	nobleChainCfg := noble.Config()

	fiatOwner, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "fiat-owner", "leg stove oblige forest occur range jar observe ahead morning street forward amazing negative digital syrup bar doctor fortune purpose buddy quote laptop civil", 1, noble)
	require.NoError(t, err)

	fiatPauser, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "fiat-pauser", "", 1, noble)
	require.NoError(t, err)

	fiatMasterMinter, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "fiat-master-minter", "", 1, noble)
	require.NoError(t, err)

	fiatMinterController, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "fiat-minter-controller", "", 1, noble)
	require.NoError(t, err)

	cctpAttesterManager, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "cctp-attester-manager", "", 1, noble)
	require.NoError(t, err)

	cctpTokenController, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "cctp-token-controller", "", 1, noble)
	require.NoError(t, err)

	cctpPauser, err := interchaintest.GetAndFundTestUserWithMnemonic(ctx, "cctp-pauser", "", 1, noble)
	require.NoError(t, err)

	err = testutil.WaitForBlocks(ctx, 2, noble)
	require.NoError(t, err)

	val := noble.Validators[0]

	// keysToRestore := []ibc.Wallet{fiatOwner, fiatMasterMinter, fiatMinterController, fiatPauser}
	// for _, wallet := range keysToRestore {
	// 	val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic())
	// }

	_, err = val.ExecTx(ctx, fiatOwner.KeyName(),
		"fiat-tokenfactory", "update-pauser", fiatPauser.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to update pauser")

	_, err = val.ExecTx(ctx, fiatPauser.KeyName(), "fiat-tokenfactory", "unpause")
	require.NoError(t, err, "failed to set fiat-tokenfactory paused state")

	_, err = val.ExecTx(ctx, paramAuthority.KeyName(), "cctp", "update-attester-manager", cctpAttesterManager.FormattedAddress())
	require.NoError(t, err, "error updating attester manager")

	_, err = val.ExecTx(ctx, paramAuthority.KeyName(), "cctp", "update-token-controller", cctpTokenController.FormattedAddress())
	require.NoError(t, err, "error updating token controller")

	_, err = val.ExecTx(ctx, paramAuthority.KeyName(), "cctp", "update-pauser", cctpPauser.FormattedAddress())
	require.NoError(t, err, "error updating pauser")

	queryRolesResults, _, err := val.ExecQuery(ctx, "cctp", "roles")
	require.NoError(t, err, "error querying cctp roles")

	var cctpRoles cctptypes.QueryRolesResponse
	err = json.Unmarshal(queryRolesResults, &cctpRoles)
	require.NoError(t, err, "failed to unmarshall cctp roles")

	// For CI testing purposes, the paramauthority and cctp owner are the same.
	require.Equal(t, paramAuthority.FormattedAddress(), cctpRoles.Owner)
	require.Equal(t, cctpAttesterManager.FormattedAddress(), cctpRoles.AttesterManager)
	require.Equal(t, cctpTokenController.FormattedAddress(), cctpRoles.TokenController)
	require.Equal(t, cctpPauser.FormattedAddress(), cctpRoles.Pauser)

	_, err = val.ExecTx(ctx, paramAuthority.KeyName(), "cctp", "update-max-message-body-size", "9000")
	require.NoError(t, err, "error updating max message body size")

	queryMaxMsgBodySize, _, err := val.ExecQuery(ctx, "cctp", "show-max-message-body-size")
	require.NoError(t, err, "error querying cctp max message body size")

	t.Logf("Max message body size: %s", string(queryMaxMsgBodySize))

	// err = json.Unmarshal(queryMaxMsgBodySize, &maxMsgBodySize)
	// require.NoError(t, err, "failed to unmarshall max message body size")

	// require.Equal(t, uint64(500), maxMsgBodySize.Amount.Amount)

	attesters := make([]*ecdsa.PrivateKey, 2)
	msgs := make([]sdk.Msg, 2)

	// attester - ECDSA public key (Circle will own these keys for mainnet)
	for i := range attesters {
		p, err := crypto.GenerateKey() // private key
		require.NoError(t, err)

		attesters[i] = p

		pubKey := elliptic.Marshal(p.PublicKey, p.PublicKey.X, p.PublicKey.Y) //public key

		attesterPub := hex.EncodeToString(pubKey)

		// Adding an attester to protocol
		msgs[i] = &cctptypes.MsgEnableAttester{
			From:     cctpAttesterManager.FormattedAddress(),
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

	tokenMessenger := make([]byte, 32)
	copy(tokenMessenger[12:], common.FromHex("0xBd3fa81B58Ba92a82136038B25aDec7066af3155"))

	bCtx, bCancel := context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err := cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		cctpAttesterManager,
		msgs...,
	)
	require.NoError(t, err, "error submitting add public keys tx")
	require.Zero(t, tx.Code, "cctp add pub keys transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("Submitted add public keys tx: %s", tx.TxHash)

	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		cctpTokenController,
		// maps remote token on remote domain to a local token -- used for minting
		&cctptypes.MsgLinkTokenPair{
			From:         cctpTokenController.FormattedAddress(),
			RemoteDomain: 0,
			RemoteToken:  burnToken,
			LocalToken:   denomMetadataUsdc.Base,
		},
	)
	require.NoError(t, err, "error submitting add token pair tx")
	require.Zero(t, tx.Code, "cctp add token pair transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("Submitted add token pair tx: %s", tx.TxHash)

	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		paramAuthority,
		&cctptypes.MsgAddRemoteTokenMessenger{
			From:     paramAuthority.FormattedAddress(),
			DomainId: 0,
			Address:  tokenMessenger,
		},
	)
	require.NoError(t, err, "error submitting add remote token messenger tx")
	require.Zero(t, tx.Code, "cctp add remote token messenger transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("Submitted add remote token messenger tx: %s", tx.TxHash)

	cctpModuleAccount := authtypes.NewModuleAddress(cctptypes.ModuleName).String()

	// by using mock images `mock-v2.0.0` or `mock-v0.4.2`, we have access to the fiat-tokenfactory owner accout
	_, err = val.ExecTx(ctx, fiatOwner.KeyName(),
		"fiat-tokenfactory", "update-master-minter", fiatMasterMinter.FormattedAddress(), "-b", "block")
	require.NoError(t, err, "failed to execute update master minter tx")

	_, err = val.ExecTx(ctx, fiatMasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", fiatMinterController.FormattedAddress(), cctpModuleAccount, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = val.ExecTx(ctx, fiatMinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", cctpModuleAccount, "1000000"+denomMetadataUsdc.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	const receiver = "9B6CA0C13EB603EF207C4657E1E619EF531A4D27" //account

	receiverBz, err := hex.DecodeString(receiver)
	require.NoError(t, err)

	nobleReceiver, err := bech32.ConvertAndEncode(nobleChainCfg.Bech32Prefix, receiverBz)
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

	const destinationCallerKeyName = "destination-caller"
	destinationCallerUser := interchaintest.GetAndFundTestUsers(t, ctx, destinationCallerKeyName, 1, noble)

	destinationCaller := make([]byte, 32)
	copy(destinationCaller[12:], destinationCallerUser[0].Address())

	wrappedDepositForBurn := cctptypes.Message{
		Version:           0,
		SourceDomain:      0,
		DestinationDomain: 4, // Noble is 4
		Nonce:             0, // dif per message
		Sender:            tokenMessenger,
		Recipient:         cctptypes.PaddedModuleAddress,
		DestinationCaller: destinationCaller,
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

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()
	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		destinationCallerUser[0],
		&cctptypes.MsgReceiveMessage{ //note: all messages that go to noble go through MsgReceiveMessage
			From:        destinationCallerUser[0].FormattedAddress(),
			Message:     wrappedDepositForBurnBz,
			Attestation: attestationBurn,
		},
	)
	require.NoError(t, err, "error submitting cctp burn recv tx")
	require.Zerof(t, tx.Code, "cctp burn recv transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP burn message successfully received: %s", tx.TxHash)

	balance, err := noble.GetBalance(ctx, nobleReceiver, denomMetadataUsdc.Base)
	require.NoError(t, err)

	require.Equal(t, int64(1000000), balance)
}
