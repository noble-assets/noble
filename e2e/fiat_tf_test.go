package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

func TestFiatTFUpdateBlacklister(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Update Blacklister while TF is paused
	// EXPECTED: Success; blacklister updated

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-1", math.OneInt(), noble)
	newBlacklister1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err := showBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse := fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: string(newBlacklister1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Update Blacklister from non owner account
	// EXPECTED: Request fails; blacklister not updated
	// Status:
	// 	Blacklister: newBlacklister1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newBlacklister2 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	showBlacklisterRes, err = showBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	// ACTION: Update Blacklister from blacklisted owner account
	// EXPECTED: Success; blacklister updated
	// Status:
	// 	Blacklister: newBlacklister1

	blacklistAccount(t, ctx, val, newBlacklister1, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err = showBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: string(newBlacklister2.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	unblacklistAccount(t, ctx, val, newBlacklister2, nw.fiatTfRoles.Owner)

	// ACTION: Update Blacklister to blacklisted Blacklister account
	// EXPECTED: Success; blacklister updated
	// Status:
	// 	Blacklister: newBlacklister2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-3", math.OneInt(), noble)
	newBlacklister3 := w[0]

	blacklistAccount(t, ctx, val, newBlacklister2, newBlacklister3)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister3.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err = showBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: string(newBlacklister3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)
}

func TestFiatTFBlacklist(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Blacklist user while TF is paused
	// EXPECTED: Success; user blacklisted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "to-blacklist-1", math.OneInt(), noble)
	toBlacklist1 := w[0]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, toBlacklist1)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Blacklist user from non Blacklister account
	// EXPECTED: Request failed; user not blacklisted

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	toBlacklist2 := w[0]
	alice := w[1]

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	var preFailedBlacklist, postFailedBlacklist fiattokenfactorytypes.QueryAllBlacklistedResponse
	_ = json.Unmarshal(res, &preFailedBlacklist)
	// ignore the error since `pagination` does not unmarshal)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the blacklister: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")
	_ = json.Unmarshal(res, &postFailedBlacklist)
	// ignore the error since `pagination` does not unmarshal)
	require.ElementsMatch(t, preFailedBlacklist.Blacklisted, postFailedBlacklist.Blacklisted)

	// Blacklist an account while the blacklister is blacklisted
	// EXPECTED: Success; user blacklisted
	// Status:
	// 	blacklisted: toBlacklist1
	// 	not blacklisted: toBlacklist2

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Blacklister)

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, toBlacklist2)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Blacklister)

	// Blacklist an already blacklisted account
	// EXPECTED: Request fails; user remains blacklisted
	// Status:
	// 	blacklisted: toBlacklist1, toBlacklist2

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Blacklister.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist1.FormattedAddress())
	require.ErrorContains(t, err, "user is already blacklisted")

	showBlacklistedRes, err := showBlacklisted(ctx, val, toBlacklist1)
	require.NoError(t, err, "failed to query show-blacklisted")
	expectedBlacklistResponse := fiattokenfactorytypes.QueryGetBlacklistedResponse{
		Blacklisted: fiattokenfactorytypes.Blacklisted{
			AddressBz: toBlacklist1.Address(),
		},
	}
	require.Equal(t, expectedBlacklistResponse.Blacklisted, showBlacklistedRes.Blacklisted)
}

func TestFiatTFUnblacklist(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Unblacklist user while TF is paused
	// EXPECTED: Success; user unblacklisted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "blacklist-user-1", math.OneInt(), noble)
	blacklistedUser1 := w[0]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, blacklistedUser1)

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, blacklistedUser1)

	// ACTION: Unblacklist user from non Blacklister account
	// EXPECTED: Request fails; user not unblacklisted
	// Status:
	// 	not blacklisted: blacklistedUser1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, blacklistedUser1)

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")
	var preFailedUnblacklist, postFailedUnblacklist fiattokenfactorytypes.QueryAllBlacklistedResponse
	_ = json.Unmarshal(res, &preFailedUnblacklist)
	// ignore the error since `pagination` does not unmarshal)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "unblacklist", blacklistedUser1.FormattedAddress())
	require.ErrorContains(t, err, "you are not the blacklister: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")
	_ = json.Unmarshal(res, &postFailedUnblacklist)
	// ignore the error since `pagination` does not unmarshal)
	require.ElementsMatch(t, preFailedUnblacklist.Blacklisted, postFailedUnblacklist.Blacklisted)

	// ACTION: Unblacklist an account while the blacklister is blacklisted
	// EXPECTED: Success; user unblacklisted
	// Status:
	// 	blacklisted: blacklistedUser1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Blacklister)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, blacklistedUser1)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Blacklister)

	// ACTION: Unblacklist an account that is not blacklisted
	// EXPECTED: Request fails; user remains unblacklisted
	// Status:
	// 	not blacklisted: blacklistedUser1

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", blacklistedUser1.FormattedAddress())
	require.ErrorContains(t, err, "the specified address is not blacklisted")

	_, err = showBlacklisted(ctx, val, blacklistedUser1)
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
}
