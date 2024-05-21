package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
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

	// ACTION: Accept owner while TF is paused
	// EXPECTED: Success; pending owner accepted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", math.OneInt(), noble)
	newOwner1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-owner")
	require.NoError(t, err, "failed to query show-owner")
	var showOwnerResponse fiattokenfactorytypes.QueryGetOwnerResponse
	err = json.Unmarshal(res, &showOwnerResponse)
	require.NoError(t, err, "failed to unmarshall show-owner response")
	expectedOwnerResponse := fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: string(newOwner1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerResponse.Owner)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Accept owner from non pending owner
	// EXPECTED: Request fails; pending owner not accepted
	// Status:
	// 	Owner: newOwner1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-2", math.OneInt(), noble)
	newOwner2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "update-owner", newOwner2.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.ErrorContains(t, err, "you are not the pending owner: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-owner")
	require.NoError(t, err, "failed to query show-owner")
	err = json.Unmarshal(res, &showOwnerResponse)
	require.NoError(t, err, "failed to unmarshall show-owner response")
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerResponse.Owner)

	// ACTION: Accept owner from blacklisted pending owner
	// EXPECTED: Success; pending owner accepted
	// Status:
	// 	Owner: newOwner1
	// 	Pending: newOwner2

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, newOwner2)

	_, err = val.ExecTx(ctx, newOwner2.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-owner")
	require.NoError(t, err, "failed to query show-owner")
	err = json.Unmarshal(res, &showOwnerResponse)
	require.NoError(t, err, "failed to unmarshall show-owner response")
	expectedOwnerResponse = fiattokenfactorytypes.QueryGetOwnerResponse{
		Owner: fiattokenfactorytypes.Owner{
			Address: string(newOwner2.FormattedAddress()),
		},
	}
	require.Equal(t, expectedOwnerResponse.Owner, showOwnerResponse.Owner)

}

func TestFiatTFUpdateMasterMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Update Master Minter while TF is paused
	// EXPECTED: Success; Master Minter updated

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-masterMinter-1", math.OneInt(), noble)
	newMM1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-master-minter")
	require.NoError(t, err, "failed to query show-master-minter")
	var getMasterMinterResponse fiattokenfactorytypes.QueryGetMasterMinterResponse
	err = json.Unmarshal(res, &getMasterMinterResponse)
	require.NoError(t, err, "failed to unmarshall show-master-minter response")

	expectedGetMasterMinterResponse := fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: string(newMM1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, getMasterMinterResponse.MasterMinter)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Update Master Minter from non owner account
	// EXPECTED: Request fails; Master Minter not updated
	// Status:
	// 	Master Minter: newMM1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-mm-2", math.OneInt(), noble)
	newMM2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-master-minter")
	require.NoError(t, err, "failed to query show-master-minter")
	err = json.Unmarshal(res, &getMasterMinterResponse)
	require.NoError(t, err, "failed to unmarshall show-master-minter response")
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, getMasterMinterResponse.MasterMinter)

	// ACTION: Update Master Minter from blacklisted owner account
	// EXPECTED: Success; Master Minter updated
	// Status:
	// 	Master Minter: newMM1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-master-minter")
	require.NoError(t, err, "failed to query show-master-minter")

	err = json.Unmarshal(res, &getMasterMinterResponse)
	require.NoError(t, err, "failed to unmarshall show-master-minter response")

	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: string(newMM2.FormattedAddress()),
		},
	}

	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, getMasterMinterResponse.MasterMinter)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	// ACTION: Update Master Minter to blacklisted Master Minter account
	// EXPECTED: Success; Master Minter updated
	// Status:
	// 	Master Minter: newMM2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-mm-3", math.OneInt(), noble)
	newMM3 := w[0]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, newMM3)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM3.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-master-minter")
	require.NoError(t, err, "failed to query show-master-minter")

	err = json.Unmarshal(res, &getMasterMinterResponse)
	require.NoError(t, err, "failed to unmarshall show-master-minter response")

	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: string(newMM3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, getMasterMinterResponse.MasterMinter)
}

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

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-pauser")
	require.NoError(t, err, "failed to query show-pauser")
	var getPauserResponse fiattokenfactorytypes.QueryGetPauserResponse
	err = json.Unmarshal(res, &getPauserResponse)
	require.NoError(t, err, "failed to unmarshall show-pauser response")
	expectedGetPauserResponse := fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: string(newPauser1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, getPauserResponse.Pauser)

	unpauseFiatTF(t, ctx, val, newPauser1)

	// ACTION: Update Pauser from non owner account
	// EXPECTED: Request fails; pauser not updated
	// Status:
	// 	Pauser: newPauser1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-2", math.OneInt(), noble)
	newPauser2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-pauser")
	require.NoError(t, err, "failed to query show-pauser")
	err = json.Unmarshal(res, &getPauserResponse)
	require.NoError(t, err, "failed to unmarshall show-pauser response")
	require.Equal(t, expectedGetPauserResponse.Pauser, getPauserResponse.Pauser)

	// ACTION: Update Pauser from blacklisted owner account
	// EXPECTED: Success; pauser updated
	// Status:
	// 	Pauser: newPauser1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-pauser")
	require.NoError(t, err, "failed to query show-pauser")
	err = json.Unmarshal(res, &getPauserResponse)
	require.NoError(t, err, "failed to unmarshall show-pauser response")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: string(newPauser2.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, getPauserResponse.Pauser)

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

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-pauser")
	require.NoError(t, err, "failed to query show-pauser")
	err = json.Unmarshal(res, &getPauserResponse)
	require.NoError(t, err, "failed to unmarshall show-pauser response")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: string(newPauser3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, getPauserResponse.Pauser)
}

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

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
	require.NoError(t, err, "failed to query show-blacklister")
	var getBlacklisterResponse fiattokenfactorytypes.QueryGetBlacklisterResponse
	err = json.Unmarshal(res, &getBlacklisterResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklister response")
	expectedGetBlacklisterResponse := fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: string(newBlacklister1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, getBlacklisterResponse.Blacklister)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Update Blacklister from non owner account
	// EXPECTED: Request fails; blacklister not updated
	// Status:
	// 	Blacklister: newBlacklister1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-2", math.OneInt(), noble)
	newBlacklister2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
	require.NoError(t, err, "failed to query show-blacklister")
	err = json.Unmarshal(res, &getBlacklisterResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklister response")
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, getBlacklisterResponse.Blacklister)

	// ACTION: Update Blacklister from blacklisted owner account
	// EXPECTED: Success; blacklister updated
	// Status:
	// 	Blacklister: newBlacklister1

	blacklistAccount(t, ctx, val, newBlacklister1, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
	require.NoError(t, err, "failed to query show-blacklister")
	err = json.Unmarshal(res, &getBlacklisterResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklister response")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: string(newBlacklister2.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, getBlacklisterResponse.Blacklister)

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

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
	require.NoError(t, err, "failed to query show-blacklister")
	err = json.Unmarshal(res, &getBlacklisterResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklister response")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: string(newBlacklister3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, getBlacklisterResponse.Blacklister)
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

	w = interchaintest.GetAndFundTestUsers(t, ctx, "to-blacklist-2", math.OneInt(), noble)
	toBlacklist2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	var preFailedBlacklist, postFailedBlacklist fiattokenfactorytypes.QueryAllBlacklistedResponse
	_ = json.Unmarshal(res, &preFailedBlacklist)
	// ignore the error since `pagination` does not unmarshall)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the blacklister: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")
	_ = json.Unmarshal(res, &postFailedBlacklist)
	// ignore the error since `pagination` does not unmarshall)
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

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", toBlacklist1.FormattedAddress())
	require.NoError(t, err, "failed to query show-blacklisted")
	var showBlacklistedResponse fiattokenfactorytypes.QueryGetBlacklistedResponse
	err = json.Unmarshal(res, &showBlacklistedResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklisted response")
	expectedBlacklistResponse := fiattokenfactorytypes.QueryGetBlacklistedResponse{
		Blacklisted: fiattokenfactorytypes.Blacklisted{
			AddressBz: toBlacklist1.Address(),
		},
	}
	require.Equal(t, expectedBlacklistResponse.Blacklisted, showBlacklistedResponse.Blacklisted)
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
	// ignore the error since `pagination` does not unmarshall)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "unblacklist", blacklistedUser1.FormattedAddress())
	require.ErrorContains(t, err, "you are not the blacklister: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")
	_ = json.Unmarshal(res, &postFailedUnblacklist)
	// ignore the error since `pagination` does not unmarshall)
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

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", blacklistedUser1.FormattedAddress())
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
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

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")
	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err)
	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: false,
		},
	}
	require.Equal(t, expectedPaused, showPausedResponse)

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

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")
	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err)
	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: true,
		},
	}
	require.Equal(t, expectedPaused, showPausedResponse)

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

func TestFiatTFConfigureMinterController(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Configure Minter Controller while TF is paused
	// EXPECTED: Success; Minter Controller is configured with Minter

	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", math.OneInt(), noble)
	minterController1 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", math.OneInt(), noble)
	minter1 := w[0]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")
	var showMinterController fiattokenfactorytypes.QueryGetMinterControllerResponse
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Configure Minter Controller from non Master Minter account
	// EXPECTED: Request fails; Minter Controller not configured with Minter
	// Status:
	// 	minterController1 -> minter1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-2", math.OneInt(), noble)
	minterController2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-2", math.OneInt(), noble)
	minter2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController2.FormattedAddress(), minter2.FormattedAddress())
	require.ErrorContains(t, err, "error configuring minter controller")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController2.FormattedAddress())
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	// ACTION: Configure a blacklisted Minter Controller and Minter from blacklisted Master Minter account
	// EXPECTED: Success; Minter Controller is configured with Minter
	// Status:
	// 	minterController1 -> minter1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)
	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, minterController2)
	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, minter2)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController2.FormattedAddress(), minter2.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController2.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err, "failed to unmarshall")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter2.FormattedAddress(),
			Controller: minterController2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)

	// ACTION: Configure an already configured Minter Controller with a new Minter
	// EXPECTED: Success; Minter Controller is configured with Minter. The old minter should be disascociated
	// from Minter Controller but keep its status and allowance
	// Status:
	// 	minterController1 -> minter1
	// 	minterController2 -> minter2

	// configuring minter1 to ensure allownace stays the same after assiging mintercontroller1 a new minter
	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error configuring minter controller")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	var getMintersPreUpdateMinterController, getMintersPostUpdateMinterController fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &getMintersPreUpdateMinterController)
	require.NoError(t, err, "failed to unmarshall")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.OneInt(),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, getMintersPreUpdateMinterController.Minters, "configured minter and or allowance is not as expected")

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-3", math.OneInt(), noble)
	minter3 := w[0]

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter3.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err, "failed to unmarshall")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController, "expected minter and minter controller is not as expected")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &getMintersPostUpdateMinterController)
	require.NoError(t, err, "failed to unmarshall")
	require.Equal(t, getMintersPreUpdateMinterController.Minters, getMintersPostUpdateMinterController.Minters, "the minter should not have changed since updating the minter controller with a new minter")

	// ACTION:- Configure an already configured Minter to another Minter Controller
	// EXPECTED: Success; Minter Controller is configured with new Minter. Minter can have multiple Minter Controllers.
	// Status:
	// 	minterController1 -> minter3
	// 	minterController2 -> minter2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-3", math.OneInt(), noble)
	minterController3 := w[0]

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController3.FormattedAddress(), minter3.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController3.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err, "failed to unmarshall")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController3.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err, "failed to unmarshall")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-minter-controller")
	require.NoError(t, err, "failed to query list-minter-controller")

	var listMinterController fiattokenfactorytypes.QueryAllMinterControllerResponse
	_ = json.Unmarshal(res, &listMinterController)
	// ignore error because `pagination` does not unmarshall

	expectedListMinterController := fiattokenfactorytypes.QueryAllMinterControllerResponse{
		MinterController: []fiattokenfactorytypes.MinterController{
			{ // this minter and controller were created/assigned at genesis
				Minter:     nw.fiatTfRoles.MasterMinter.FormattedAddress(),
				Controller: nw.fiatTfRoles.Minter.FormattedAddress(),
			},
			{
				Minter:     minter3.FormattedAddress(),
				Controller: minterController1.FormattedAddress(),
			},
			{
				Minter:     minter2.FormattedAddress(),
				Controller: minterController2.FormattedAddress(),
			},
			{
				Minter:     minter3.FormattedAddress(),
				Controller: minterController3.FormattedAddress(),
			},
		},
	}

	require.ElementsMatch(t, expectedListMinterController.MinterController, listMinterController.MinterController)
}

func TestFiatTFRemoveMinterController(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Remove Minter Controller while TF is paused
	// EXPECTED: Success; Minter Controller is removed

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Remove a Minter Controller from non Master Minter account
	// EXPECTED: Request fails; Minter Controller remains configured with Minter

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress(), nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")

	var showMinterController fiattokenfactorytypes.QueryGetMinterControllerResponse
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)

	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     nw.fiatTfRoles.Minter.FormattedAddress(),
			Controller: nw.fiatTfRoles.MinterController.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.ErrorContains(t, err, "you are not the master minter: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")

	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)

	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     nw.fiatTfRoles.Minter.FormattedAddress(),
			Controller: nw.fiatTfRoles.MinterController.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	// ACTION: Remove Minter Controller while Minter and Minter Controller are blacklisted
	// EXPECTED: Success; Minter Controller is removed
	// Status:
	// 	gw minterController -> gw minter

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)
	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MinterController)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)

	// ACTION: Remove a a non existent Minter Controller
	// EXPECTED: Requst fails
	// Status:
	// 	no minterController setup

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.ErrorContains(t, err, fmt.Sprintf("minter controller with a given address (%s) doesn't exist: user not found", nw.fiatTfRoles.MinterController.FormattedAddress()))

}

func TestFiatTFConfigureMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Configure minter while TF is paused
	// EXPECTED: Request fails; Minter is not configured

	// configure new minter controller and minter
	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", math.OneInt(), noble)
	minterController1 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", math.OneInt(), noble)
	minter1 := w[0]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")

	var showMinterController fiattokenfactorytypes.QueryGetMinterControllerResponse
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)

	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	allowance := int64(10)

	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.ErrorContains(t, err, "minting is paused")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.Error(t, err, "minter found; configuring minter should not have succeeded")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Configure minter from a minter controller not associated with the minter
	// EXPECTED: Request fails; Minter is not configured with new Minter Controller but old Minter retains its allowance
	// Status:
	// minterController1 -> minter1 (un-configured)
	// gw minterController -> gw minter

	// reconfigure minter to ensure balance does not change
	configureMinter(t, ctx, val, minterController1, minter1, allowance)

	differentAllowance := allowance + 99
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", differentAllowance))
	require.ErrorContains(t, err, "minter address ≠ minter controller's minter address")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	var getMintersPreUpdateMinterController fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &getMintersPreUpdateMinterController)
	require.NoError(t, err, "failed to unmarshall")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, getMintersPreUpdateMinterController.Minters, "configured minter allowance is not as expected")

	// ACTION: Configure a minter that is blacklisted
	// EXPECTED: Success; Minter is configured with allowance
	// Status:
	// minterController1 -> minter1
	// gw minterController -> gw minter

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, minter1)

	configureMinter(t, ctx, val, minterController1, minter1, 11)

}

func TestFiatTFRemoveMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Remove minter while TF is paused
	// EXPECTED: Success; Minter is removed

	allowance := int64(10)

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.Error(t, err, "minter found; not successfully removed")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Remove minter from a minter controller not associated with the minter
	// EXPECTED: Request fails; Minter is not removed
	// Status:
	// 	gw minterController -> gw minter (Removed)

	// reconfigure minter
	configureMinter(t, ctx, val, nw.fiatTfRoles.MinterController, nw.fiatTfRoles.Minter, allowance)

	minter1, _ := setupMinterAndController(t, ctx, noble, val, nw.fiatTfRoles.MasterMinter, allowance)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", minter1.FormattedAddress())
	require.ErrorContains(t, err, "minter address ≠ minter controller's minter address")

	// ensure minter still exists
	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")

	var showMinterResponse fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	// ACTION: Remove minter from a blacklisted minter controller
	// EXPECTED: Success; Minter is removed
	// Status:
	// 	gw minterController -> gw minter
	// 	minterController1 -> minter1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MinterController)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.Error(t, err, "minter found; not successfully removed")

}

func TestFiatTFMint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Mint while TF is paused
	// EXPECTED: Request fails; amount not minted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "receiver-1", math.OneInt(), noble)
	receiver1 := w[0]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	var preMintShowMinterResponse, showMinterResponse fiattokenfactorytypes.QueryGetMintersResponse

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &preMintShowMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	preMintAllowance := preMintShowMinterResponse.Minters.Allowance.Amount

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "minting is paused")

	bal, err := noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: nw.fiatTfRoles.Minter.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: preMintAllowance,
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Mint from non minter
	// EXPECTED: Request fails; amount not minted

	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "you are not a minter")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// ACTION: Mint from blacklisted minter
	// EXPECTED: Request fails; amount not minted

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "minter address is blacklisted")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	// ACTION: Mint to blacklisted account
	// EXPECTED: Request fails; amount not minted

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver1)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "receiver address is blacklisted")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver1)

	// ACTION: Mint an amount that exceeds the minters allowance
	// EXPECTED: Request fails; amount not minted

	exceedAllowance := preMintAllowance.Add(math.NewInt(99))
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", exceedAllowance.Int64()))
	require.ErrorContains(t, err, "minting amount is greater than the allowance")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	// ACTION: Successfully mint into an account
	// EXPECTED: Success

	mintAmount := int64(3)
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.Equal(math.NewInt(mintAmount)))

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters.Minters.Allowance = sdktypes.Coin{
		Denom:  "uusdc",
		Amount: preMintAllowance.Sub(math.NewInt(mintAmount)),
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

}

func TestFiatTFBurn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// setup - mint into minter's wallet
	mintAmount := int64(5)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nw.fiatTfRoles.Minter.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err := noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64())

	// ACTION: Burn while TF is paused
	// EXPECTED: Request fails; amount not burned

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	burnAmount := int64(1)
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "burning is paused")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Burn from non minter account
	// EXPECTED: Request fails; amount not burned

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.NewInt(burnAmount), noble)
	alice := w[0]

	// mint into Alice's account to give her a balance to burn
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err = noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.Equal(math.NewInt(mintAmount)))

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "you are not a minter")

	bal, err = noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	// ACTION: Burn from a blacklisted minter account
	// EXPECTED: Request fails; amount not burned

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "minter address is blacklisted")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Minter)

	// ACTION: Burn amount greater than the minters balance
	// EXPECTED: Request fails; amount not burned

	exceedAllowance := bal.Add(math.NewInt(99))
	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", exceedAllowance.Int64()))
	require.ErrorContains(t, err, "insufficient funds")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	// ACTION: Successfully burn tokens
	// EXPECTED: Success; amount burned and Minters balance is decreased

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	bal, err = noble.GetBalance(ctx, nw.fiatTfRoles.Minter.FormattedAddress(), "uusdc")
	expectedAmount := mintAmount - burnAmount
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, expectedAmount, bal.Int64())
}

func TestFiatTFAuthzGrant(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// ACTION: Grant an authz SEND using a TF token while TF is paused
	// EXPECTED: ??????

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter1 := w[0]
	grantee1 := w[1]

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	res, err := val.AuthzGrant(ctx, granter1, grantee1.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Grant an authz SEND using a TF token to a grantee who is blacklisted
	// EXPECTED: Success;

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter2 := w[0]
	grantee2 := w[1]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, grantee2)

	res, err = val.AuthzGrant(ctx, granter2, grantee2.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	// ACTION: Grant an authz SEND using a TF token from a granter who is blacklisted
	// EXPECTED: Success;

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter3 := w[0]
	grantee3 := w[1]

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, granter3)

	res, err = val.AuthzGrant(ctx, granter3, grantee3.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)
}

func TestFiatTFAuthzSend(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// setup
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble, noble)
	granter := w[0]
	grantee := w[1]
	receiver := w[2]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", granter.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	res, err := val.AuthzGrant(ctx, granter, grantee.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	sendAmount := 5
	nestedCmd := []string{
		noble.Config().Bin,
		"tx", "bank", "send", granter.FormattedAddress(), receiver.FormattedAddress(), fmt.Sprintf("%duusdc", sendAmount),
		"--from", granter.FormattedAddress(), "--generate-only",
		"--chain-id", noble.GetNode().Chain.Config().ChainID,
		"--node", noble.GetNode().Chain.GetRPCAddress(),
		"--home", noble.GetNode().HomeDir(),
		"--keyring-backend", keyring.BackendTest,
		"--output", "json",
		"--yes",
	}

	// ACTION: Execute an authz SEND using a TF token from a grantee who is blacklisted
	// EXPECTED: Succeed; Grantee is acting on behalf of Granter
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, grantee)

	res, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.NoError(t, err)
	require.Zero(t, res.Code)

	bal, err := noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.EqualValues(t, sendAmount, bal.Int64())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, grantee)

	// ACTION: Execute an authz SEND using a TF token from a granter who is blacklisted
	// EXPECTED: Request fails; Granter is blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, granter)

	preSendBal := bal

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens: unauthorized", granter.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	// bal should not change
	require.Equal(t, preSendBal, bal)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, granter)

	// ACTION: Execute an authz SEND using a TF token to a receiver who is blacklisted
	// EXPECTED: Request fails; Granter is blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver)

	preSendBal = bal

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens: unauthorized", receiver.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	// bal should not change
	require.Equal(t, preSendBal, bal)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, receiver)

	// ACTION: Execute an authz SEND using a TF token while the TF is paused
	// EXPECTED: Request fails; chain is paused
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	preSendBal = bal

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, "the chain is paused")

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	// bal should not change
	require.Equal(t, preSendBal, bal)

}

func TestFiatTFBankSend(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw := nobleSpinUp(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	alice := w[0]
	bob := w[1]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// ACTION: Send TF token while TF is paused
	// EXPECTED: Request fails; token not sent

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	amountToSend := ibc.WalletAmount{
		Address: bob.FormattedAddress(),
		Denom:   "uusdc",
		Amount:  math.OneInt(),
	}
	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, "the chain is paused")

	bobBal, err := noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Send TF token while FROM address is blacklisted
	// EXPECTED: Request fails; token not sent

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, alice)

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens", alice.FormattedAddress()))

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, alice)

	// ACTION: Send TF token while TO address is blacklisted
	// EXPECTED: Request fails; token not sent

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, bob)

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens", bob.FormattedAddress()))

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, bob)

	// ACTION: Successfully send TF token
	// EXPECTED: Success

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.NoError(t, err, "error sending funds")

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, amountToSend.Amount, bobBal)
}

func TestFiatTFIBCOut(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, gaia, r, ibcPathName, eRep := nobleSpinUpIBC(t, ctx, true)
	noble := nw.chain
	val := noble.Validators[0]

	// setup
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, gaia)
	nobleWallet := w[0]
	gaiaWallet := w[1]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.fiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nobleWallet.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// noble -> gaia channel info
	nobleToGaiaChannelInfo, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)
	nobleToGaiaChannelID := nobleToGaiaChannelInfo[0].ChannelID

	// gaia -> noble channel info
	gaiaToNobleChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaToNobleChannelID := gaiaToNobleChannelInfo[0].ChannelID

	amountToSend := math.NewInt(5)
	transfer := ibc.WalletAmount{
		Address: gaiaWallet.FormattedAddress(),
		Denom:   denomMetadataUsdc.Base,
		Amount:  amountToSend,
	}

	// ACTION: IBC send TF token while TF is paused
	// EXPECTED: Request fails;

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ibc transfer noble -> gaia
	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, "the chain is paused")

	// relay MsgRecvPacket & MsgAcknowledgement
	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	// uusdc IBC denom on gaia
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaToNobleChannelID, denomMetadataUsdc.Base))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	gaiaWalletBal, err := gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: IBC send TF token from a blacklisted user
	// EXPECTED: Request fails;

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleWallet)

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens", nobleWallet.FormattedAddress()))

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleWallet)

	// ACTION: IBC send TF token to a blacklisted user
	// EXPECTED: Request fails;

	_, bz, err := bech32.DecodeAndConvert(gaiaWallet.FormattedAddress())
	require.NoError(t, err)
	require.NoError(t, sdktypes.VerifyAddressFormat(bz))

	nobleBechOfGaiaWal := cosmos.NewWallet("default", bz, gaiaWallet.Mnemonic(), noble.Config())
	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nobleBechOfGaiaWal)

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens", gaiaWallet.FormattedAddress()))

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

}

// blacklistAccount blacklists an account and then runs the `show-blacklisted` query to ensure the
// account was successfully blacklisted on chain
func blacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, toBlacklist ibc.Wallet) {
	_, err := val.ExecTx(ctx, blacklister.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", toBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to query show-blacklisted")

	var showBlacklistedResponse fiattokenfactorytypes.QueryGetBlacklistedResponse
	err = json.Unmarshal(res, &showBlacklistedResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklisted response")

	expectedBlacklistResponse := fiattokenfactorytypes.QueryGetBlacklistedResponse{
		Blacklisted: fiattokenfactorytypes.Blacklisted{
			AddressBz: toBlacklist.Address(),
		},
	}

	require.Equal(t, expectedBlacklistResponse.Blacklisted, showBlacklistedResponse.Blacklisted)
}

// unblacklistAccount unblacklists an account and then runs the `show-blacklisted` query to ensure the
// account was successfully unblacklisted on chain
func unblacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, unBlacklist ibc.Wallet) {
	_, err := val.ExecTx(ctx, blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", unBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", unBlacklist.FormattedAddress())
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
}

// pauseFiatTF pauses the fiat tokenfactory. It then runs the `show-paused` query to ensure the
// the tokenfactory was successfully paused
func pauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser ibc.Wallet) {
	_, err := val.ExecTx(ctx, pauser.KeyName(), "fiat-tokenfactory", "pause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")

	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err)

	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: true,
		},
	}
	require.Equal(t, expectedPaused, showPausedResponse)
}

// unpauseFiatTF pauses the fiat tokenfactory. It then runs the `show-paused` query to ensure the
// the tokenfactory was successfully unpaused
func unpauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser ibc.Wallet) {
	_, err := val.ExecTx(ctx, pauser.KeyName(), "fiat-tokenfactory", "unpause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")

	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err, "failed to unmarshall show-puased response")

	expectedUnpaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: false,
		},
	}
	require.Equal(t, expectedUnpaused, showPausedResponse)
}

// setupMinterAndController creates a minter controller and minter. It also sets up a minter with an specified allowance of `uusdc`
func setupMinterAndController(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, val *cosmos.ChainNode, masterMinter ibc.Wallet, allowance int64) (minter ibc.Wallet, minterController ibc.Wallet) {
	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", math.OneInt(), noble)
	minterController = w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", math.OneInt(), noble)
	minter = w[0]

	_, err := val.ExecTx(ctx, masterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController.FormattedAddress(), minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")

	var showMinterController fiattokenfactorytypes.QueryGetMinterControllerResponse
	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)

	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter.FormattedAddress(),
			Controller: minterController.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	configureMinter(t, ctx, val, minterController, minter, allowance)

	return minter, minterController
}

// configureMinter configures a minter with a specified allowance of `uusdc`. It then runs the `show-minters` query to ensure
// the minter was properly configured
func configureMinter(t *testing.T, ctx context.Context, val *cosmos.ChainNode, minterController, minter ibc.Wallet, allowance int64) {
	_, err := val.ExecTx(ctx, minterController.KeyName(), "fiat-tokenfactory", "configure-minter", minter.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.NoError(t, err, "error configuring minter")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")

	var showMinterResponse fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)
}

// nobleSpinUp starts noble chain
// Args:
//
//	setupAllFiatTFRoles: if true, all Tokenfactory roles will be created and setup at genesis,
//		if false, only the Onwer role will be created
func nobleSpinUp(t *testing.T, ctx context.Context, setupAllFiatTFRoles bool) (nw nobleWrapper) {
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	numValidators := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &nw, "noble-1", numValidators, numFullNodes, setupAllFiatTFRoles),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	nw.chain = chains[0].(*cosmos.CosmosChain)
	noble := nw.chain

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	return
}

// nobleSpinUpIBC is the same as nobleSpinUp except it also spins up gaia chain and creates
// an IBC path between them
// Args:
//
//	setupAllFiatTFRoles: if true, all Tokenfactory roles will be created and setup at genesis,
//		if false, only the Onwer role will be created
func nobleSpinUpIBC(t *testing.T, ctx context.Context, setupAllFiatTFRoles bool) (nw nobleWrapper, gaia *cosmos.CosmosChain, r ibc.Relayer, ibcPathName string, eRep *testreporter.RelayerExecReporter) {
	rep := testreporter.NewNopReporter()
	eRep = rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	numValidators := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &nw, "noble-1", numValidators, numFullNodes, setupAllFiatTFRoles),
		{Name: "gaia", Version: "latest", NumValidators: &numValidators, NumFullNodes: &numFullNodes},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	nw.chain = chains[0].(*cosmos.CosmosChain)
	noble := nw.chain
	gaia = chains[1].(*cosmos.CosmosChain)

	rf := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t))
	r = rf.Build(t, client, network)

	ibcPathName = "path"
	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: r,
			Path:    ibcPathName,
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

	return
}
