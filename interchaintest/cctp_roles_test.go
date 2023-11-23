package interchaintest_test

import (
	"context"
	"testing"

	"github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestCCTP_UpdateOwner(t *testing.T) {
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

	noble, _, interchain, _ := SetupInterchain(t, ctx, logger, execReporter, client, network, &wrapper, TokenFactoryConfiguration{
		false, false, false, false,
	})

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	var err error
	nobleValidator := noble.Validators[0]

	//

	currentOwner := wrapper.fiatTfRoles.Owner
	newOwner := wrapper.extraWallets.User

	_, err = nobleValidator.ExecTx(ctx, currentOwner.KeyName(),
		"cctp", "update-owner", newOwner.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update owner tx")

	roles, err := getRoles(nobleValidator, ctx)
	require.NoError(t, err, "failed to query roles")
	require.Equal(t, currentOwner.FormattedAddress(), roles.Owner)

	_, err = nobleValidator.ExecTx(ctx, newOwner.KeyName(),
		"cctp", "accept-owner",
	)
	require.NoError(t, err, "failed to execute accept owner tx")

	roles, err = getRoles(nobleValidator, ctx)
	require.NoError(t, err, "failed to query roles")
	require.Equal(t, newOwner.FormattedAddress(), roles.Owner)
}

func TestCCTP_UpdateAttesterManager(t *testing.T) {
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

	noble, _, interchain, _ := SetupInterchain(t, ctx, logger, execReporter, client, network, &wrapper, TokenFactoryConfiguration{
		false, false, false, false,
	})

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	var err error
	nobleValidator := noble.Validators[0]

	//

	currentAttesterManager := wrapper.fiatTfRoles.Owner
	newAttesterManager := wrapper.extraWallets.User

	_, err = nobleValidator.ExecTx(ctx, currentAttesterManager.KeyName(),
		"cctp", "update-attester-manager", newAttesterManager.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update attester manager tx")

	roles, err := getRoles(nobleValidator, ctx)
	require.NoError(t, err, "failed to query roles")
	require.Equal(t, newAttesterManager.FormattedAddress(), roles.AttesterManager)
}

func TestCCTP_UpdatePauser(t *testing.T) {
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

	noble, _, interchain, _ := SetupInterchain(t, ctx, logger, execReporter, client, network, &wrapper, TokenFactoryConfiguration{
		false, false, false, false,
	})

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	var err error
	nobleValidator := noble.Validators[0]

	//

	currentPauser := wrapper.fiatTfRoles.Owner
	newPauser := wrapper.extraWallets.User

	_, err = nobleValidator.ExecTx(ctx, currentPauser.KeyName(),
		"cctp", "update-pauser", newPauser.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update pauser tx")

	roles, err := getRoles(nobleValidator, ctx)
	require.NoError(t, err, "failed to query roles")
	require.Equal(t, newPauser.FormattedAddress(), roles.Pauser)
}

func TestCCTP_UpdateTokenController(t *testing.T) {
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

	noble, _, interchain, _ := SetupInterchain(t, ctx, logger, execReporter, client, network, &wrapper, TokenFactoryConfiguration{
		false, false, false, false,
	})

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	var err error
	nobleValidator := noble.Validators[0]

	//

	currentTokenController := wrapper.fiatTfRoles.Owner
	newTokenController := wrapper.extraWallets.User

	_, err = nobleValidator.ExecTx(ctx, currentTokenController.KeyName(),
		"cctp", "update-token-controller", newTokenController.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update token controller tx")

	roles, err := getRoles(nobleValidator, ctx)
	require.NoError(t, err, "failed to query roles")
	require.Equal(t, newTokenController.FormattedAddress(), roles.TokenController)
}

func getRoles(validator *cosmos.ChainNode, ctx context.Context) (roles types.QueryRolesResponse, err error) {
	res, _, err := validator.ExecQuery(ctx, "cctp", "roles")
	if err != nil {
		return
	}

	err = jsonpb.UnmarshalString(string(res), &roles)
	return
}
