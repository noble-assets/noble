package e2e_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func TestCCTP_UpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw := e2e.NobleSpinUp(t, ctx, true)
	noble := nw.Chain
	nobleValidator := noble.Validators[0]

	cctpOwner := nw.CCTPRoles.Owner
	newOwner := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.OneInt(), noble)[0]

	_, err := nobleValidator.ExecTx(ctx, cctpOwner.KeyName(),
		"cctp", "update-owner", newOwner.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute update owner tx")

	roles, err := getRoles(nobleValidator, ctx)
	require.NoError(t, err, "failed to query roles")
	require.Equal(t, cctpOwner.FormattedAddress(), roles.Owner)

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
	nw := e2e.NobleSpinUp(t, ctx, true)
	noble := nw.Chain
	nobleValidator := noble.Validators[0]

	cctpOwner := nw.CCTPRoles.Owner
	newAttesterManager := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.OneInt(), noble)[0]

	_, err := nobleValidator.ExecTx(ctx, cctpOwner.KeyName(),
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
	nw := e2e.NobleSpinUp(t, ctx, true)
	noble := nw.Chain
	nobleValidator := noble.Validators[0]

	cctpOwner := nw.CCTPRoles.Owner
	newPauser := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.OneInt(), noble)[0]

	_, err := nobleValidator.ExecTx(ctx, cctpOwner.KeyName(),
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
	nw := e2e.NobleSpinUp(t, ctx, true)
	noble := nw.Chain
	nobleValidator := noble.Validators[0]

	cctpOwner := nw.CCTPRoles.Owner
	newTokenController := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.OneInt(), noble)[0]

	_, err := nobleValidator.ExecTx(ctx, cctpOwner.KeyName(),
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
