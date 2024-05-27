package e2e

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

func TestFiatTFUpdatePauser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Update Pauser while TF is paused
	// EXPECTED: Success; pauser updated

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-1", math.OneInt(), noble)
	newPauser1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err := showPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse := fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: string(newPauser1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	unpauseFiatTF(t, ctx, val, newPauser1)

	// ACTION: Update Pauser from non owner account
	// EXPECTED: Request fails; pauser not updated
	// Status:
	// 	Pauser: newPauser1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newPauser2 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	showPauserRes, err = showPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	// ACTION: Update Pauser from blacklisted owner account
	// EXPECTED: Success; pauser updated
	// Status:
	// 	Pauser: newPauser1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err = showPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: string(newPauser2.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	// ACTION: Update Pauser to blacklisted Pauser account
	// EXPECTED: Success; pauser updated
	// Status:
	// 	Pauser: newPauser2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-3", math.OneInt(), noble)
	newPauser3 := w[0]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, newPauser3)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser3.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err = showPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: string(newPauser3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)
}

func TestFiatTFPause(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Pause TF from an account that is not the Pauser
	// EXPECTED: Request fails; TF not paused

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err := val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "pause")
	require.ErrorContains(t, err, "you are not the pauser: unauthorized")

	showPausedRes, err := showPaused(ctx, val)
	require.NoError(t, err, "error querying for paused state")
	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: false,
		},
	}
	require.Equal(t, expectedPaused, showPausedRes)

	// ACTION: Pause TF from a blacklisted Pauser account
	// EXPECTED: Success; TF is paused
	// Status:
	// 	Paused: false

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Pauser)

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Pauser)

	// ACTION: Pause TF while TF is already paused
	// EXPECTED: Success; TF remains paused
	// Status:
	// 	Paused: true

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)
}

func TestFiatTFUnpause(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Unpause TF from an account that is not a Pauser
	// EXPECTED: Request fails; TF remains paused

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err := val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "unpause")
	require.ErrorContains(t, err, "you are not the pauser: unauthorized")

	showPausedRes, err := showPaused(ctx, val)
	require.NoError(t, err, "error querying for paused state")
	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: true,
		},
	}
	require.Equal(t, expectedPaused, showPausedRes)

	// ACTION: Unpause TF from a blacklisted Pauser account
	// EXPECTED: Success; TF is unpaused
	// Status:
	// 	Paused: true

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Pauser)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Pauser)

	// ACTION: Unpause TF while TF is already paused
	// EXPECTED: Success; TF remains unpaused
	// Status:
	// 	Paused: false

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)
}
