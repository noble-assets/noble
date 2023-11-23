package interchaintest_test

import (
	"context"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestICS20BPSFees(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	reporter := testreporter.NewNopReporter()
	execReporter := reporter.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	var wrapper genesisWrapper

	noble, gaia, interchain, rly := SetupInterchain(t, ctx, logger, execReporter, client, network, &wrapper, TokenFactoryConfiguration{
		false, false, true, false,
	})

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	var err error
	var res ibc.RelayerExecResult
	nobleValidator := noble.Validators[0]

	_, err = nobleValidator.ExecTx(ctx, wrapper.fiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", wrapper.fiatTfRoles.MinterController.FormattedAddress(), wrapper.fiatTfRoles.Minter.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, wrapper.fiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", wrapper.fiatTfRoles.Minter.FormattedAddress(), "1000000000000"+denomMetadataUsdc.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, wrapper.fiatTfRoles.Minter.KeyName(),
		"fiat-tokenfactory", "mint", wrapper.extraWallets.User.FormattedAddress(), "1000000000000"+denomMetadataUsdc.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, wrapper.extraWallets.User.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equalf(t, int64(1000000000000), userBalance, "failed to mint %s to user", denomMetadataUsdc.Base)

	nobleChans, err := rly.GetChannels(ctx, execReporter, noble.Config().ChainID)
	require.NoError(t, err, "failed to get noble channels")
	require.Len(t, nobleChans, 2, "more than two channels found")
	nobleChan := nobleChans[1]

	gaiaReceiver := "cosmos169xaqmxumqa829gg73nxrenkhhd2mrs36j3vrz"

	// First, test BPS below max fees
	_, err = noble.SendIBCTransfer(ctx, nobleChan.ChannelID, wrapper.extraWallets.User.KeyName(), ibc.WalletAmount{
		Address: gaiaReceiver,
		Denom:   denomMetadataUsdc.Base,
		Amount:  100000000,
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer from noble")

	res = rly.Exec(ctx, execReporter, []string{"hermes", "clear", "packets", "--chain", noble.Config().ChainID, "--port", "transfer", "--channel", nobleChan.ChannelID}, nil)
	require.NoError(t, res.Err)

	userBalance, err = noble.GetBalance(ctx, wrapper.extraWallets.User.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(999900000000), userBalance, "user balance is incorrect")

	prefixedDenom := transfertypes.GetPrefixedDenom(nobleChan.Counterparty.PortID, nobleChan.Counterparty.ChannelID, denomMetadataUsdc.Base)
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	ibcDenom := denomTrace.IBCDenom()

	// 100000000 (Transfer Amount) * .0001 (1 BPS) = 10000 taken as fees
	receiverBalance, err := gaia.GetBalance(ctx, gaiaReceiver, ibcDenom)
	require.NoError(t, err, "failed to get receiver balance")
	require.Equal(t, int64(99990000), receiverBalance, "receiver balance incorrect")

	// of the 10000 taken as fees, 75% goes to distribution entity (7500)
	distributionEntityBalance, err := noble.GetBalance(ctx, wrapper.paramAuthority.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get distribution entity balance")
	require.Equal(t, int64(7500), distributionEntityBalance, "distribution entity balance incorrect")

	// Now test max fee
	_, err = noble.SendIBCTransfer(ctx, nobleChan.ChannelID, wrapper.extraWallets.User.FormattedAddress(), ibc.WalletAmount{
		Address: gaiaReceiver,
		Denom:   denomMetadataUsdc.Base,
		Amount:  100000000000,
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer from noble")

	res = rly.Exec(ctx, execReporter, []string{"hermes", "clear", "packets", "--chain", noble.Config().ChainID, "--port", "transfer", "--channel", nobleChan.ChannelID}, nil)
	require.NoError(t, res.Err)

	// 999900000000 user balance from prior test, now subtract 100000000000 = 899900000000
	userBalance, err = noble.GetBalance(ctx, wrapper.extraWallets.User.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(899900000000), userBalance, "user balance is incorrect")

	// fees will max, 5000000 is taken off of transfer amount
	// prior receiver balance 99990000. add 100000000000 transfer amount but subtracted 5000000 in bps fees (max) = 100094990000
	receiverBalance, err = gaia.GetBalance(ctx, gaiaReceiver, ibcDenom)
	require.NoError(t, err, "failed to get receiver balance")
	require.Equal(t, int64(100094990000), receiverBalance, "receiver balance incorrect")

	// prior balance 7500, add 75% of the 5000000 fee (3750000) = 3757500
	distributionEntityBalance, err = noble.GetBalance(ctx, wrapper.paramAuthority.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get distribution entity balance")
	require.Equal(t, int64(3757500), distributionEntityBalance, "distribution entity balance incorrect")
}
