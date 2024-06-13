package e2e

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

func TestFiatTFUpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Update owner while TF is paused
	// EXPECTED: Success; Pending owner set

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", math.OneInt(), noble)
	newOwner1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Update owner from unprivileged account
	// EXPECTED: Request fails; pending owner not set

	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	// ACTION: Update Owner from blacklisted owner account
	// EXPECTED: Success; pending owner set

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	// ACTION: Update Owner to a blacklisted account
	// EXPECTED: Success; pending owner set

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, newOwner1)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

}

func TestFiatTFAcceptOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Happy path: accept owner
	// EXPECTED: Success; pending owner accepted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", math.OneInt(), noble)
	newOwner1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	showOwnerRes, err := showOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	expectedOwnerResponse := fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: string(newOwner1.FormattedAddress()),
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

	showOwnerRes, err = showOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	// ACTION: Accept owner while TF is paused
	// EXPECTED: Success; pending owner accepted
	// Status:
	// 	Owner: newOwner1

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "update-owner", newOwner2.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, newOwner2.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	showOwnerRes, err = showOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	expectedOwnerResponse = fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: string(newOwner2.FormattedAddress()),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

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

	showOwnerRes, err = showOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

	// ACTION: Accept owner from blacklisted pending owner
	// EXPECTED: Success; pending owner accepted
	// Status:
	// 	Owner: newOwner2
	// 	Pending: newOwner3

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, newOwner3)

	_, err = val.ExecTx(ctx, newOwner3.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	showOwnerRes, err = showOwner(ctx, val)
	require.NoError(t, err, "failed to query show-owner")
	expectedOwnerResponse = fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: string(newOwner3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerRes.Owner)

}
