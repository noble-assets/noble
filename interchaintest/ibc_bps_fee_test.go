package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/interchaintest/v3/testutil"
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

	var (
		noble, gaia          *cosmos.CosmosChain
		roles, roles2        NobleRoles
		extraWallets         ExtraWallets
		paramauthorityWallet ibc.Wallet
	)

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "token",
		Bech32Prefix:   "noble",
		CoinType:       "118",
		GasPrices:      "0.0token",
		GasAdjustment:  1.1,
		TrustingPeriod: "504h",
		NoHostMount:    false,
		Images:         nobleImageInfo,
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) (err error) {
			val := noble.Validators[0]
			err = createTokenfactoryRoles(ctx, &roles, denomMetadataRupee, val, false)
			if err != nil {
				return err
			}
			err = createTokenfactoryRoles(ctx, &roles2, denomMetadataDrachma, val, false)
			if err != nil {
				return err
			}
			extraWallets, err = createExtraWalletsAtGenesis(ctx, val)
			if err != nil {
				return err
			}
			paramauthorityWallet, err = createParamAuthAtGenesis(ctx, val)
			return err
		},
		ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			g := make(map[string]interface{})
			if err := json.Unmarshal(b, &g); err != nil {
				return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
			}
			if err := modifyGenesisTokenfactory(g, "tokenfactory", denomMetadataRupee, &roles, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisTokenfactory(g, "fiat-tokenfactory", denomMetadataDrachma, &roles2, false); err != nil {
				return nil, err
			}
			if err := modifyGenesisParamAuthority(g, paramauthorityWallet.FormattedAddress()); err != nil {
				return nil, err
			}
			if err := modifyGenesisTariffDefaults(g, paramauthorityWallet.FormattedAddress()); err != nil {
				return nil, err
			}
			out, err := json.Marshal(&g)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
			}
			return out, nil
		},
	}

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			ChainConfig:   chainCfg,
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
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

	noble, gaia = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)
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

	_, err = nobleValidator.ExecTx(ctx, roles2.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", roles2.MinterController.FormattedAddress(), roles2.Minter.FormattedAddress(), "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, roles2.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", roles2.Minter.FormattedAddress(), "1000000000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, roles2.Minter.KeyName(),
		"fiat-tokenfactory", "mint", extraWallets.User.FormattedAddress(), "1000000000000"+denomMetadataDrachma.Base, "-b", "block",
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	userBalance, err := noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equalf(t, int64(1000000000000), userBalance, "failed to mint %s to user", denomMetadataDrachma.Base)

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
	tx, err := noble.SendIBCTransfer(ctx, nobleChan.ChannelID, extraWallets.User.KeyName(), ibc.WalletAmount{
		Address: gaiaReceiver,
		Denom:   denomMetadataDrachma.Base,
		Amount:  100000000,
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer from noble")

	_, err = testutil.PollForAck(ctx, noble, height, height+10, tx.Packet)
	require.NoError(t, err, "failed to find ack for ibc transfer")

	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(999900000000), userBalance, "user balance is incorrect")

	prefixedDenom := transfertypes.GetPrefixedDenom(nobleChan.Counterparty.PortID, nobleChan.Counterparty.ChannelID, denomMetadataDrachma.Base)
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	ibcDenom := denomTrace.IBCDenom()

	// 100000000 (Transfer Amount) * .0001 (1 BPS) = 10000 taken as fees
	receiverBalance, err := gaia.GetBalance(ctx, gaiaReceiver, ibcDenom)
	require.NoError(t, err, "failed to get receiver balance")
	require.Equal(t, int64(99990000), receiverBalance, "receiver balance incorrect")

	// of the 10000 taken as fees, 80% goes to distribution entity (8000)
	distributionEntityBalance, err := noble.GetBalance(ctx, paramauthorityWallet.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err, "failed to get distribution entity balance")
	require.Equal(t, int64(8000), distributionEntityBalance, "distribution entity balance incorrect")

	// Now test max fee
	tx, err = noble.SendIBCTransfer(ctx, nobleChan.ChannelID, extraWallets.User.FormattedAddress(), ibc.WalletAmount{
		Address: gaiaReceiver,
		Denom:   denomMetadataDrachma.Base,
		Amount:  100000000000,
	}, ibc.TransferOptions{})
	require.NoError(t, err, "failed to send ibc transfer from noble")

	_, err = testutil.PollForAck(ctx, noble, height, height+10, tx.Packet)
	require.NoError(t, err, "failed to find ack for ibc transfer")

	// 999900000000 user balance from prior test, now subtract 100000000000 = 899900000000
	userBalance, err = noble.GetBalance(ctx, extraWallets.User.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err, "failed to get user balance")
	require.Equal(t, int64(899900000000), userBalance, "user balance is incorrect")

	// fees will max, 5000000 is taken off of transfer amount
	// prior receiver balance 99990000. add 100000000000 transfer amount but subtracted 5000000 in bps fees (max) = 100094990000
	receiverBalance, err = gaia.GetBalance(ctx, gaiaReceiver, ibcDenom)
	require.NoError(t, err, "failed to get receiver balance")
	require.Equal(t, int64(100094990000), receiverBalance, "receiver balance incorrect")

	// prior balance 8000, add 80% of the 5000000 fee (4000000) = 4008000
	distributionEntityBalance, err = noble.GetBalance(ctx, paramauthorityWallet.FormattedAddress(), denomMetadataDrachma.Base)
	require.NoError(t, err, "failed to get distribution entity balance")
	require.Equal(t, int64(4008000), distributionEntityBalance, "distribution entity balance incorrect")

}
