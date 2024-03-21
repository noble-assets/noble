package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/icza/dyno"
	forwardingtypes "github.com/noble-assets/noble/v5/x/forwarding/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/relayer/hermes"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestForwarding_RegisterOnNoble(t *testing.T) {
	t.Parallel()

	ctx, wrapper, gaia, _, _, sender, receiver := ForwardingSuite(t)
	validator := wrapper.chain.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
	require.True(t, exists)

	require.NoError(t, validator.SendFunds(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  1_000_000,
	}))
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, wrapper.chain, gaia))

	senderBalance, err := wrapper.chain.AllBalances(ctx, sender.FormattedAddress())
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err := wrapper.chain.AllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, int64(1_000_000), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", sdk.NewInt(1_000_000))), stats.TotalForwarded)
}

func TestForwarding_RegisterViaTransfer(t *testing.T) {
	t.Parallel()

	ctx, wrapper, gaia, _, _, _, receiver := ForwardingSuite(t)
	validator := wrapper.chain.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	_, err := gaia.SendIBCTransfer(ctx, "channel-0", receiver.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uatom",
		Amount:  100_000,
	}, ibc.TransferOptions{
		Memo: fmt.Sprintf("{\"noble\":{\"forwarding\":{\"recipient\":\"%s\"}}}", receiver.FormattedAddress()),
	})
	require.NoError(t, err)

	require.NoError(t, testutil.WaitForBlocks(ctx, 10, wrapper.chain, gaia))

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
	require.True(t, exists)

	balance, err := wrapper.chain.AllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), "uatom")
	require.NoError(t, err)
	require.Equal(t, int64(998_000), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uatom",
	}.IBCDenom(), sdk.NewInt(100_000))), stats.TotalForwarded)
}

func TestForwarding_RegisterViaPacket(t *testing.T) {
	t.Skip()
}

func TestForwarding_FrontRunAccount(t *testing.T) {
	t.Parallel()

	ctx, wrapper, gaia, _, _, sender, receiver := ForwardingSuite(t)
	validator := wrapper.chain.Validators[0]

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	require.NoError(t, validator.SendFunds(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  1_000_000,
	}))

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
	require.True(t, exists)

	require.NoError(t, testutil.WaitForBlocks(ctx, 10, wrapper.chain, gaia))

	senderBalance, err := wrapper.chain.AllBalances(ctx, sender.FormattedAddress())
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err := wrapper.chain.AllBalances(ctx, address)
	require.NoError(t, err)
	require.True(t, balance.IsZero())

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, int64(1_000_000), receiverBalance)

	stats := ForwardingStats(t, ctx, validator)
	require.Equal(t, uint64(1), stats.NumOfAccounts)
	require.Equal(t, uint64(1), stats.NumOfForwards)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("uusdc", sdk.NewInt(1_000_000))), stats.TotalForwarded)
}

func TestForwarding_ClearAccount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx, wrapper, gaia, rly, execReporter, sender, receiver := ForwardingSuite(t)
	validator := wrapper.chain.Validators[0]

	require.NoError(t, rly.StopRelayer(ctx, execReporter))

	address, exists := ForwardingAccount(t, ctx, validator, receiver)
	require.False(t, exists)

	_, err := validator.ExecTx(ctx, sender.KeyName(), "forwarding", "register-account", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	_, exists = ForwardingAccount(t, ctx, validator, receiver)
	require.True(t, exists)

	require.NoError(t, validator.SendFunds(ctx, sender.KeyName(), ibc.WalletAmount{
		Address: address,
		Denom:   "uusdc",
		Amount:  1_000_000,
	}))

	time.Sleep(10 * time.Minute)

	require.NoError(t, rly.StartRelayer(ctx, execReporter))
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, wrapper.chain, gaia))

	senderBalance, err := wrapper.chain.AllBalances(ctx, sender.FormattedAddress())
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err := wrapper.chain.GetBalance(ctx, address, "uusdc")
	require.NoError(t, err)
	require.Equal(t, int64(1_000_000), balance)

	receiverBalance, err := gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, int64(0), receiverBalance)

	_, err = validator.ExecTx(ctx, sender.KeyName(), "forwarding", "clear-account", address)
	require.NoError(t, err)
	require.NoError(t, testutil.WaitForBlocks(ctx, 10, wrapper.chain, gaia))

	senderBalance, err = wrapper.chain.AllBalances(ctx, sender.FormattedAddress())
	require.NoError(t, err)
	require.True(t, senderBalance.IsZero())

	balance, err = wrapper.chain.GetBalance(ctx, address, "uusdc")
	require.NoError(t, err)
	require.Equal(t, int64(0), balance)

	receiverBalance, err = gaia.GetBalance(ctx, receiver.FormattedAddress(), transfertypes.DenomTrace{
		Path:      "transfer/channel-0",
		BaseDenom: "uusdc",
	}.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, int64(1_000_000), receiverBalance)
}

//

func ForwardingAccount(t *testing.T, ctx context.Context, validator *cosmos.ChainNode, receiver ibc.Wallet) (address string, exists bool) {
	raw, _, err := validator.ExecQuery(ctx, "forwarding", "address", "channel-0", receiver.FormattedAddress())
	require.NoError(t, err)

	var res forwardingtypes.QueryAddressResponse
	require.NoError(t, json.Unmarshal(raw, &res))

	return res.Address, res.Exists
}

func ForwardingStats(t *testing.T, ctx context.Context, validator *cosmos.ChainNode) forwardingtypes.QueryStatsByChannelResponse {
	raw, _, err := validator.ExecQuery(ctx, "forwarding", "stats", "channel-0")
	require.NoError(t, err)

	var res forwardingtypes.QueryStatsByChannelResponse
	require.NoError(t, jsonpb.UnmarshalString(string(raw), &res))

	return res
}

func ForwardingSuite(t *testing.T) (ctx context.Context, wrapper genesisWrapper, gaia *cosmos.CosmosChain, rly *hermes.Relayer, execReporter *testreporter.RelayerExecReporter, sender ibc.Wallet, receiver ibc.Wallet) {
	ctx = context.Background()
	logger := zaptest.NewLogger(t)
	reporter := testreporter.NewNopReporter()
	execReporter = reporter.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	numValidators, numFullNodes := 1, 0

	spec := nobleChainSpec(ctx, &wrapper, "noble-1", numValidators, numFullNodes, false, false, false, false)
	spec.ModifyGenesis = func(cfg ibc.ChainConfig, bz []byte) ([]byte, error) {
		bz, err := modifyGenesisAll(&wrapper, false, false)(cfg, bz)
		if err != nil {
			return nil, err
		}

		genesis := make(map[string]interface{})
		if err := json.Unmarshal(bz, &genesis); err != nil {
			return nil, err
		}

		if err := dyno.Set(genesis, "0", "app_state", "tariff", "params", "transfer_fee_bps"); err != nil {
			return nil, err
		}
		if err := dyno.Set(genesis, "0", "app_state", "tariff", "params", "transfer_fee_max"); err != nil {
			return nil, err
		}

		return json.Marshal(&genesis)
	}

	factory := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		spec,
		{
			Name:          "gaia",
			Version:       "v14.1.0",
			NumValidators: &numValidators,
			NumFullNodes:  &numFullNodes,
			ChainConfig: ibc.ChainConfig{
				ChainID: "cosmoshub-4",
			},
		},
	})

	chains, err := factory.Chains(t.Name())
	require.NoError(t, err)

	noble := chains[0].(*cosmos.CosmosChain)
	gaia = chains[1].(*cosmos.CosmosChain)
	wrapper.chain = noble

	rly = interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		logger,
	).Build(t, client, network).(*hermes.Relayer)

	interchain := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(rly, "rly").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: rly,
			Path:    "transfer",
		})

	require.NoError(t, interchain.Build(ctx, execReporter, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
	}))

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	require.NoError(t, rly.StartRelayer(ctx, execReporter))

	roles := wrapper.fiatTfRoles
	sender = wrapper.extraWallets.User
	validator := noble.Validators[0]

	_, err = validator.ExecTx(ctx, roles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", roles.MinterController.FormattedAddress(), roles.Minter.FormattedAddress(), "-b", "block")
	require.NoError(t, err)
	_, err = validator.ExecTx(ctx, roles.MinterController.KeyName(), "fiat-tokenfactory", "configure-minter", roles.Minter.FormattedAddress(), "1000000uusdc", "-b", "block")
	require.NoError(t, err)
	_, err = validator.ExecTx(ctx, roles.Minter.KeyName(), "fiat-tokenfactory", "mint", sender.FormattedAddress(), "1000000uusdc", "-b", "block")
	require.NoError(t, err)

	receivers := interchaintest.GetAndFundTestUsers(t, ctx, "receiver", 1_000_000, gaia)
	receiver = receivers[0]

	return
}
