package interchaintest_test

import (
	"context"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
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

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var gw genesisWrapper

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", nv, nf, false, false, true, false),
		{
			Name:          "gaia",
			Version:       "v9.0.2",
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayerImage,
	).Build(t, client, network)

	var gaia *cosmos.CosmosChain
	gw.chain, gaia = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)
	noble := gw.chain

	path := "p"

	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Path:    path,
			Relayer: r,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	nobleValidator := noble.Validators[0]

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", gw.fiatTfRoles.MinterController.FormattedAddress(), gw.fiatTfRoles.Minter.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", gw.fiatTfRoles.Minter.FormattedAddress(), "1000000000000"+denomMetadataUsdc.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, gw.fiatTfRoles.Minter.KeyName(),
		"fiat-tokenfactory", "mint", gw.extraWallets.User.FormattedAddress(), "1000000000000"+denomMetadataUsdc.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, gw.extraWallets.User.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equalf(t, int64(1000000000000), userBalance, "failed to mint %s to user", denomMetadataUsdc.Base)

	nobleChans, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err, "failed to get noble channels")
	require.Len(t, nobleChans, 1, "more than one channel found")
	nobleChan := nobleChans[0]

	gaiaReceiver := "cosmos169xaqmxumqa829gg73nxrenkhhd2mrs36j3vrz"

	err = r.StartRelayer(ctx, eRep, path)
	require.NoError(t, err, "failed to start relayer")
	defer r.StopRelayer(ctx, eRep)

	height, err := noble.Height(ctx)
	require.NoError(t, err, "failed to get noble height")

	// First, test BPS below max fees
	tx, err := noble.SendIBCTransfer(ctx, nobleChan.ChannelID, gw.extraWallets.User.KeyName(), ibc.WalletAmount{
		Address: gaiaReceiver,
		Denom:   denomMetadataUsdc.Base,
		Amount:  100000000,
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer from noble")

	_, err = testutil.PollForAck(ctx, noble, height, height+10, tx.Packet)
	require.NoError(t, err, "failed to find ack for ibc transfer")

	userBalance, err = noble.GetBalance(ctx, gw.extraWallets.User.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(999900000000), userBalance, "user balance is incorrect")

	prefixedDenom := transfertypes.GetPrefixedDenom(nobleChan.Counterparty.PortID, nobleChan.Counterparty.ChannelID, denomMetadataUsdc.Base)
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	ibcDenom := denomTrace.IBCDenom()

	// 100000000 (Transfer Amount) * .0001 (1 BPS) = 10000 taken as fees
	receiverBalance, err := gaia.GetBalance(ctx, gaiaReceiver, ibcDenom)
	require.NoError(t, err, "failed to get receiver balance")
	require.Equal(t, int64(99990000), receiverBalance, "receiver balance incorrect")

	// of the 10000 taken as fees, 80% goes to distribution entity (8000)
	distributionEntityBalance, err := noble.GetBalance(ctx, gw.paramAuthority.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get distribution entity balance")
	require.Equal(t, int64(8000), distributionEntityBalance, "distribution entity balance incorrect")

	// Now test max fee
	tx, err = noble.SendIBCTransfer(ctx, nobleChan.ChannelID, gw.extraWallets.User.FormattedAddress(), ibc.WalletAmount{
		Address: gaiaReceiver,
		Denom:   denomMetadataUsdc.Base,
		Amount:  100000000000,
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer from noble")

	_, err = testutil.PollForAck(ctx, noble, height, height+10, tx.Packet)
	require.NoError(t, err, "failed to find ack for ibc transfer")

	// 999900000000 user balance from prior test, now subtract 100000000000 = 899900000000
	userBalance, err = noble.GetBalance(ctx, gw.extraWallets.User.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(899900000000), userBalance, "user balance is incorrect")

	// fees will max, 5000000 is taken off of transfer amount
	// prior receiver balance 99990000. add 100000000000 transfer amount but subtracted 5000000 in bps fees (max) = 100094990000
	receiverBalance, err = gaia.GetBalance(ctx, gaiaReceiver, ibcDenom)
	require.NoError(t, err, "failed to get receiver balance")
	require.Equal(t, int64(100094990000), receiverBalance, "receiver balance incorrect")

	// prior balance 8000, add 80% of the 5000000 fee (4000000) = 4008000
	distributionEntityBalance, err = noble.GetBalance(ctx, gw.paramAuthority.FormattedAddress(), denomMetadataUsdc.Base)
	require.NoError(t, err, "failed to get distribution entity balance")
	require.Equal(t, int64(4008000), distributionEntityBalance, "distribution entity balance incorrect")

}
