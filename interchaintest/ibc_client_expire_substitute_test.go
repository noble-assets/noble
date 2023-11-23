package interchaintest_test

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/relayer/hermes"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestClientSubstitution(t *testing.T) {
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
		true, true, true, true,
	})

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	nobleChainID := noble.Config().ChainID
	gaiaChainID := gaia.Config().ChainID

	var err error
	var res ibc.RelayerExecResult

	// create client on noble with short trusting period which will expire.
	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "client", "--host-chain", nobleChainID, "--reference-chain", gaiaChainID, "--trusting-period", "1m"}, nil)
	require.NoError(t, res.Err)

	nobleClientID, err := hermes.GetClientIdFromStdout(res.Stdout)
	require.NoError(t, err)

	// create client on gaia with longer trusting period so it won't expire for this test.
	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "client", "--host-chain", gaiaChainID, "--reference-chain", nobleChainID}, nil)
	require.NoError(t, res.Err)

	gaiaClientID, err := hermes.GetClientIdFromStdout(res.Stdout)
	require.NoError(t, err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "connection", "--a-chain", nobleChainID, "--a-client", nobleClientID, "--b-client", gaiaClientID}, nil)
	require.NoError(t, res.Err)

	nobleConnectionID, _, err := hermes.GetConnectionIDsFromStdout(res.Stdout)
	require.NoError(t, err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "channel", "--a-chain", nobleChainID, "--a-connection", nobleConnectionID, "--a-port", "transfer", "--b-port", "transfer"}, nil)
	require.NoError(t, res.Err)

	nobleChannelID, _, err := hermes.GetChannelIDsFromStdout(res.Stdout)
	require.NoError(t, err)

	const userFunds = int64(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, noble, gaia)

	err = testutil.WaitForBlocks(ctx, 20, noble)
	require.NoError(t, err)

	// client should now be expired, no relayer was running to update the clients during the 20s trusting period.

	_, err = noble.SendIBCTransfer(ctx, nobleChannelID, users[0].KeyName(), ibc.WalletAmount{
		Address: users[1].FormattedAddress(),
		Amount:  1000000,
		Denom:   noble.Config().Denom,
	}, ibc.TransferOptions{})

	require.Error(t, err)
	require.ErrorContains(t, err, "status Expired: client is not active")

	// create new client on noble
	res = rly.Exec(ctx, execReporter, []string{"hermes", "--json", "create", "client", "--host-chain", nobleChainID, "--reference-chain", gaiaChainID}, nil)
	require.NoError(t, res.Err)

	newNobleClientID, err := hermes.GetClientIdFromStdout(res.Stdout)
	require.NoError(t, err)

	// substitute new client state into old client
	_, err = noble.Validators[0].ExecTx(ctx, wrapper.paramAuthority.KeyName(), "ibc-authority", "update-client", nobleClientID, newNobleClientID)
	require.NoError(t, err)

	// send a packet on the same channel with new client, should succeed.
	_, err = noble.SendIBCTransfer(ctx, nobleChannelID, users[0].KeyName(), ibc.WalletAmount{
		Address: users[1].FormattedAddress(),
		Amount:  1000000,
		Denom:   noble.Config().Denom,
	}, ibc.TransferOptions{})
	require.NoError(t, err)

	res = rly.Exec(ctx, execReporter, []string{"hermes", "clear", "packets", "--chain", noble.Config().ChainID, "--port", "transfer", "--channel", nobleChannelID}, nil)
	require.NoError(t, res.Err)
}
