package interchaintest_test

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/noble-assets/noble/v5/cmd"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/relayer/hermes"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type TokenFactoryConfiguration struct {
	MinSetupTF      bool
	MinSetupFiatTF  bool
	MinModifyTF     bool
	MinModifyFiatTF bool
}

func SetupInterchain(t *testing.T, ctx context.Context, logger *zap.Logger, execReporter *testreporter.RelayerExecReporter, client *client.Client, network string, wrapper *genesisWrapper, config TokenFactoryConfiguration) (noble *cosmos.CosmosChain, gaia *cosmos.CosmosChain, interchain *interchaintest.Interchain, rly *hermes.Relayer) {
	factory := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, wrapper, "noble-1", 1, 0, config.MinSetupTF, config.MinSetupFiatTF, config.MinModifyTF, config.MinModifyFiatTF),
		{
			Name:    "gaia",
			Version: "latest",
			ChainConfig: ibc.ChainConfig{
				ChainID: "cosmoshub-4",
			},
		},
	})

	chains, err := factory.Chains(t.Name())
	require.NoError(t, err)

	noble = chains[0].(*cosmos.CosmosChain)
	gaia = chains[1].(*cosmos.CosmosChain)
	wrapper.chain = noble

	rly = interchaintest.NewBuiltinRelayerFactory(
		ibc.Hermes,
		logger,
	).Build(t, client, network).(*hermes.Relayer)

	interchain = interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(rly, "rly").
		AddProviderConsumerLink(interchaintest.ProviderConsumerLink{
			Provider: gaia,
			Consumer: noble,
			Relayer:  rly,
		})

	require.NoError(t, interchain.Build(ctx, execReporter, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: true,
	}))

	cmd.SetPrefixes(noble.Config().Bech32Prefix)

	var res ibc.RelayerExecResult
	nobleClientID, gaiaClientID := "07-tendermint-0", "07-tendermint-0"

	require.NoError(t, rly.MarkChainAsConsumer(ctx, noble.Config().ChainID))

	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "connection", "--a-chain", noble.Config().ChainID, "--a-client", nobleClientID, "--b-client", gaiaClientID}, nil)
	require.NoError(t, res.Err)

	nobleConnectionID, gaiaConnectionID, err := hermes.GetConnectionIDsFromStdout(res.Stdout)
	require.NoError(t, err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "channel", "--a-chain", noble.Config().ChainID, "--a-connection", nobleConnectionID, "--a-port", "consumer", "--b-port", "provider", "--order", "ORDER_ORDERED"}, nil)
	require.NoError(t, res.Err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "tx", "chan-open-try", "--dst-chain", gaia.Config().ChainID, "--src-chain", noble.Config().ChainID, "--dst-connection", gaiaConnectionID, "--dst-port", "transfer", "--src-port", "transfer", "--src-channel", "channel-1"}, nil)
	require.NoError(t, res.Err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "tx", "chan-open-ack", "--dst-chain", noble.Config().ChainID, "--src-chain", gaia.Config().ChainID, "--dst-connection", nobleConnectionID, "--dst-port", "transfer", "--src-port", "transfer", "--dst-channel", "channel-1", "--src-channel", "channel-1"}, nil)
	require.NoError(t, res.Err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "tx", "chan-open-confirm", "--dst-chain", gaia.Config().ChainID, "--src-chain", noble.Config().ChainID, "--dst-connection", gaiaConnectionID, "--dst-port", "transfer", "--src-port", "transfer", "--dst-channel", "channel-1", "--src-channel", "channel-1"}, nil)
	require.NoError(t, res.Err)

	delegators := interchaintest.GetAndFundTestUsers(t, ctx, "delegator", 1000000000000, gaia)
	delegator := delegators[0]

	validator, err := gaia.Validators[0].KeyBech32(ctx, "validator", "val")
	require.NoError(t, err)

	_, err = gaia.FullNodes[0].ExecTx(ctx, delegator.KeyName(), "staking", "delegate", validator, "999999000000uatom")
	require.NoError(t, err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "clear", "packets", "--chain", noble.Config().ChainID, "--port", "consumer", "--channel", "channel-0"}, nil)
	require.NoError(t, res.Err)
	res = rly.Exec(ctx, execReporter, []string{"hermes", "clear", "packets", "--chain", noble.Config().ChainID, "--port", "transfer", "--channel", "channel-1"}, nil)
	require.NoError(t, res.Err)

	return
}
