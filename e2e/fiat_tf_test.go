package e2e

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

func TestFiatTFIBCOut(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, gaia, r, ibcPathName, eRep := nobleSpinUpIBC(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, gaia)
	nobleWallet := w[0]
	gaiaWallet := w[1]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nobleWallet.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// noble -> gaia channel info
	nobleToGaiaChannelInfo, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)
	nobleToGaiaChannelID := nobleToGaiaChannelInfo[0].ChannelID
	// gaia -> noble channel info
	gaiaToNobleChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaToNobleChannelID := gaiaToNobleChannelInfo[0].ChannelID

	amountToSend := math.NewInt(5)
	transfer := ibc.WalletAmount{
		Address: gaiaWallet.FormattedAddress(),
		Denom:   denomMetadataUsdc.Base,
		Amount:  amountToSend,
	}

	// ACTION: IBC send out a TF token while TF is paused
	// EXPECTED: Request fails;

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ibc transfer noble -> gaia
	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, "the chain is paused")

	// relay MsgRecvPacket & MsgAcknowledgement
	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	// uusdc IBC denom on gaia
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaToNobleChannelID, denomMetadataUsdc.Base))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	gaiaWalletBal, err := gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: IBC send TF token from a blacklisted user
	// EXPECTED: Request fails;

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleWallet)

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens", nobleWallet.FormattedAddress()))

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleWallet)

	// ACTION: IBC send out a TF token to a blacklisted user
	// EXPECTED: Request fails;

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, gaiaWallet)

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens", gaiaWallet.FormattedAddress()))

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, gaiaWallet)

	// ACTION: IBC send out a TF token to malformed address
	// EXPECTED: Request fails;

	transfer.Address = "malformed-address1234"

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, "decoding bech32 failed")

	// ACTION: Successfully IBC send out a TF token to an address on another chain
	// EXPECTED: Success;

	transfer.Address = gaiaWallet.FormattedAddress()

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, gaiaWalletBal)
}

func TestFiatTFIBCIn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, gaia, r, ibcPathName, eRep := nobleSpinUpIBC(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.NewInt(1_000_000), noble, gaia)
	nobleWallet := w[0] // 1_000_000ustake
	gaiaWallet := w[1]  // 1_000_000uatom

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nobleWallet.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// noble -> gaia channel info
	nobleToGaiaChannelInfo, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)
	nobleToGaiaChannelID := nobleToGaiaChannelInfo[0].ChannelID
	// gaia -> noble channel info
	gaiaToNobleChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaToNobleChannelID := gaiaToNobleChannelInfo[0].ChannelID

	amountToSend := math.NewInt(mintAmount)
	transfer := ibc.WalletAmount{
		Address: gaiaWallet.FormattedAddress(),
		Denom:   denomMetadataUsdc.Base,
		Amount:  amountToSend,
	}

	// ibc transfer noble -> gaia
	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// relay MsgRecvPacket & MsgAcknowledgement
	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	// uusdc IBC denom on gaia
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaToNobleChannelID, denomMetadataUsdc.Base))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	gaiaWalletBal, err := gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, gaiaWalletBal)

	// ACTION: IBC send in a TF token while TF is paused
	// EXPECTED: Request fails;

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	amountToSend = math.OneInt()
	transfer = ibc.WalletAmount{
		Address: nobleWallet.FormattedAddress(),
		Denom:   dstIbcDenom,
		Amount:  amountToSend,
	}

	height, err := noble.Height(ctx)
	require.NoError(t, err)

	tx, err := gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err := noble.Height(ctx)
	require.NoError(t, err)

	ack, err := testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.Contains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err := noble.GetBalance(ctx, nobleWallet.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err)
	require.True(t, nobleWalletBal.IsZero())

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: IBC send in a TF token FROM an address that is blacklisted
	// EXPECTED: Request fails;

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, gaiaWallet)

	height, err = noble.Height(ctx)
	require.NoError(t, err)

	tx, err = gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err = noble.Height(ctx)
	require.NoError(t, err)

	ack, err = testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.Contains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err = noble.GetBalance(ctx, nobleWallet.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err)
	require.True(t, nobleWalletBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, gaiaWallet)

	// ACTION: IBC send in a TF token TO an address that is blacklisted
	// EXPECTED: Request fails;

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleWallet)

	height, err = noble.Height(ctx)
	require.NoError(t, err)

	tx, err = gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err = noble.Height(ctx)
	require.NoError(t, err)

	ack, err = testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.Contains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err = noble.GetBalance(ctx, nobleWallet.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err)
	require.True(t, nobleWalletBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleWallet)

	// ACTION: Successfully IBC send in a TF token to an address on noble
	// EXPECTED: Success;

	height, err = noble.Height(ctx)
	require.NoError(t, err)

	tx, err = gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err = noble.Height(ctx)
	require.NoError(t, err)

	ack, err = testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.NotContains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err = noble.GetBalance(ctx, nobleWallet.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, nobleWalletBal)
}
