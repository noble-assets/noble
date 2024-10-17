package e2e_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

func TestFiatTFUpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw := e2e.NobleSpinUp(t, ctx, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Update owner while TF is paused
	// EXPECTED: Success; Pending owner set

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", math.OneInt(), noble)
	newOwner1 := w[0]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Update owner from unprivileged account
	// EXPECTED: Request fails; pending owner not set

	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	// ACTION: Update Owner from blacklisted owner account
	// EXPECTED: Success; pending owner set

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	// ACTION: Update Owner to a blacklisted account
	// EXPECTED: Success; pending owner set

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, newOwner1)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")
}

func TestFiatTFAcceptOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw := e2e.NobleSpinUp(t, ctx, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: accept owner
	// EXPECTED: Success; pending owner accepted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", math.OneInt(), noble)
	newOwner1 := w[0]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	showOwnerRes, err := e2e.ShowOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	expectedOwnerResponse := fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: newOwner1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	// ACTION: Accept owner when no pending owner is set
	// EXPECTED: Request fails; pending owner not set
	// Status:
	// 	Owner: newOwner1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", math.OneInt(), noble)
	newOwner2 := w[0]

	_, err = val.ExecTx(ctx, newOwner2.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.ErrorContains(t, err, "pending owner is not set")

	showOwnerRes, err = e2e.ShowOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	// ACTION: Accept owner while TF is paused
	// EXPECTED: Success; pending owner accepted
	// Status:
	// 	Owner: newOwner1

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "update-owner", newOwner2.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, newOwner2.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	showOwnerRes, err = e2e.ShowOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	expectedOwnerResponse = fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: newOwner2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Accept owner from non pending owner
	// EXPECTED: Request fails; pending owner not accepted
	// Status:
	// 	Owner: newOwner2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newOwner3 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, newOwner2.KeyName(), "fiat-tokenfactory", "update-owner", newOwner3.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.ErrorContains(t, err, "you are not the pending owner: unauthorized")

	showOwnerRes, err = e2e.ShowOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	// ACTION: Accept owner from blacklisted pending owner
	// EXPECTED: Success; pending owner accepted
	// Status:
	// 	Owner: newOwner2
	// 	Pending: newOwner3

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, newOwner3)

	_, err = val.ExecTx(ctx, newOwner3.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	showOwnerRes, err = e2e.ShowOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	expectedOwnerResponse = fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: newOwner3.FormattedAddress(),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)
}
