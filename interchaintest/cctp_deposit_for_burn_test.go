package interchaintest_test

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

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
func TestCCTP_DepositForBurn(t *testing.T) {
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
		nobleChainSpec(ctx, &gw, "grand-1", nv, nf, false, false, true, false),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain

	cmd.SetPrefixes(noble.Config().Bech32Prefix)

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	nobleValidator := noble.Validators[0]

	// SET UP FIAT TOKEN FACTORY AND MINT

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", gw.fiatTfRoles.MinterController.FormattedAddress(), gw.fiatTfRoles.Minter.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", gw.fiatTfRoles.Minter.FormattedAddress(), "1000000000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.Minter.KeyName(),
		"fiat-tokenfactory", "mint", gw.extraWallets.User.FormattedAddress(), "1000000000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", gw.fiatTfRoles.MinterController.FormattedAddress(), cctptypes.ModuleAddress.String(), "-b", "block",
	)
	require.NoError(t, err, "failed to configure cctp minter controller")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", cctptypes.ModuleAddress.String(), "1000000000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to configure cctp minter")

	// ----

	broadcaster := cosmos.NewBroadcaster(t, noble)
	broadcaster.ConfigureClientContextOptions(func(clientContext sdkclient.Context) sdkclient.Context {
		return clientContext.WithBroadcastMode(flags.BroadcastBlock)
	})

	burnToken := make([]byte, 32)
	copy(burnToken[12:], common.FromHex("0x07865c6E87B9F70255377e024ace6630C1Eaa37F"))

	tokenMessenger := make([]byte, 32)
	copy(tokenMessenger[12:], common.FromHex("0xD0C3da58f55358142b8d3e06C1C30c5C6114EFE8"))

	msgs := []sdk.Msg{}

	msgs = append(msgs, &cctptypes.MsgAddRemoteTokenMessenger{
		From:     gw.fiatTfRoles.Owner.FormattedAddress(),
		DomainId: 0,
		Address:  tokenMessenger,
	})

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
	require.NoError(t, err, "error configuring remote domain")
	require.Zero(t, tx.Code, "configuring remote domain failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	beforeBurnBal, err := noble.GetBalance(ctx, gw.extraWallets.User.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err)

	mintRecipient := make([]byte, 32)
	copy(mintRecipient[12:], common.FromHex("0xfCE4cE85e1F74C01e0ecccd8BbC4606f83D3FC90"))

	depositForBurnNoble := &cctptypes.MsgDepositForBurn{
		From:              gw.extraWallets.User.FormattedAddress(),
		Amount:            cosmossdk_io_math.NewInt(1000000),
		BurnToken:         denomMetadataDrachma.Base,
		DestinationDomain: 0,
		MintRecipient:     mintRecipient,
	}

	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		gw.extraWallets.User,
		depositForBurnNoble,
	)
	require.NoError(t, err, "error broadcasting msgDepositForBurn")
	require.Zero(t, tx.Code, "msgDepositForBurn failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	afterBurnBal, err := noble.GetBalance(ctx, gw.extraWallets.User.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err)

	require.Equal(t, afterBurnBal, beforeBurnBal-1000000)

	for _, rawEvent := range tx.Events {
		switch rawEvent.Type {
		case "circle.cctp.v1.DepositForBurn":
			parsedEvent, err := sdk.ParseTypedEvent(rawEvent)
			require.NoError(t, err)
			depositForBurn, ok := parsedEvent.(*cctptypes.DepositForBurn)
			require.True(t, ok)

			expectedBurnToken := hex.EncodeToString(crypto.Keccak256([]byte(denomMetadataDrachma.Base)))

			require.Equal(t, uint64(0), depositForBurn.Nonce)
			require.Equal(t, expectedBurnToken, depositForBurn.BurnToken)
			require.Equal(t, depositForBurnNoble.Amount, depositForBurn.Amount)
			require.Equal(t, gw.extraWallets.User.FormattedAddress(), depositForBurn.Depositor)
			require.Equal(t, mintRecipient, depositForBurn.MintRecipient)
			require.Equal(t, uint32(0), depositForBurn.DestinationDomain)
			require.Equal(t, tokenMessenger, depositForBurn.DestinationTokenMessenger)
			require.Equal(t, []byte{}, depositForBurn.DestinationCaller)

		case "circle.cctp.v1.MessageSent":
			parsedEvent, err := sdk.ParseTypedEvent(rawEvent)
			require.NoError(t, err)
			event, ok := parsedEvent.(*cctptypes.MessageSent)
			require.True(t, ok)

			message, err := new(cctptypes.Message).Parse(event.Message)
			require.NoError(t, err)

			messageSender := make([]byte, 32)
			copy(messageSender[12:], sdk.MustAccAddressFromBech32(cctptypes.ModuleAddress.String()))

			expectedBurnToken := crypto.Keccak256([]byte(depositForBurnNoble.BurnToken))

			moduleAddress := make([]byte, 32)
			copy(moduleAddress[12:], sdk.MustAccAddressFromBech32(gw.extraWallets.User.FormattedAddress()))

			destinationCaller := make([]byte, 32)

			require.Equal(t, uint32(0), message.Version)
			require.Equal(t, uint32(4), message.SourceDomain)
			require.Equal(t, uint32(0), message.DestinationDomain)
			require.Equal(t, uint64(0), message.Nonce)
			require.Equal(t, messageSender, message.Sender)
			require.Equal(t, tokenMessenger, message.Recipient)
			require.Equal(t, destinationCaller, message.DestinationCaller)

			body, err := new(cctptypes.BurnMessage).Parse(message.MessageBody)
			require.NoError(t, err)

			require.Equal(t, uint32(0), body.Version)
			require.Equal(t, mintRecipient, body.MintRecipient)
			require.Equal(t, depositForBurnNoble.Amount, body.Amount)
			require.Equal(t, expectedBurnToken, body.BurnToken)
			require.Equal(t, moduleAddress, body.MessageSender)
		}
	}
}
