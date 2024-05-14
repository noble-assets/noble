package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TODO: after updating to latest interchaintest, replaces all tx hash queries with: noble.GetTransaction(hash)
// Note: we ignore the error when we unmarshall `txResponse` because some types do not unmarshal (ex: height of int64 vs string)
var txResponse sdktypes.TxResponse

func TestFiatTFUpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// Note: because there is no way to query for a 'pending owner', the only way
	// to ensure the 'update-owner' message succeeded is to query for the tx hash
	// and ensure a 0 response code.

	// - Update owner while paused -
	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", 1, noble)
	newOwner1 := w[0]

	hash, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", string(newOwner1.FormattedAddress()))
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Update owner from unprivileged account -
	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-2", 1, noble)
	newOwner2 := w[0]

	hash, err = val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "update-owner", string(newOwner2.FormattedAddress()))
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the owner: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	// - Update Owner from blacklisted owner account -
	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Owner)

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-3", 1, noble)
	newOwner3 := w[0]

	hash, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", string(newOwner3.FormattedAddress()))
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")

	// - Update Owner to a blacklisted account -
	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-4", 1, noble)
	newOwner4 := w[0]

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, newOwner4)

	hash, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", string(newOwner4.FormattedAddress()))
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")
}

func TestFiatTFAcceptOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Accept owner while TF is paused -

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", 1, noble)
	newOwner1 := w[0]

	hash, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-owner")
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

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Accept owner from non pending owner -
	// Status:
	// 	Owner: newOwner1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-2", 1, noble)
	newOwner2 := w[0]

	hash, err = val.ExecTx(ctx, newOwner1.KeyName(), "fiat-tokenfactory", "update-owner", newOwner2.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")

	hash, err = val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the pending owner: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	// - Accept owner from blacklisted pending owner -
	// Status:
	// 	Owner: newOwner1
	// 	Pending: newOwner2

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, newOwner2)

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

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Update Master Minter while TF is paused -

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-mm-1", 1, noble)
	newMM1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM1.FormattedAddress())
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

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Update Master Minter from non owner account -
	// Status:
	// 	Master Minter: newMM1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-mm-2", 1, noble)
	newMM2 := w[0]

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the owner: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-master-minter")
	require.NoError(t, err, "failed to query show-master-minter")

	err = json.Unmarshal(res, &getMasterMinterResponse)
	require.NoError(t, err, "failed to unmarshall show-master-minter response")

	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, getMasterMinterResponse.MasterMinter)

	// - Update Master Minter from blacklisted owner account -
	// Status:
	// 	Master Minter: newMM1

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
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

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Owner)

	// - Update Master Minter to blacklisted Master Minter account -
	// Status:
	// 	Master Minter: newMM2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-mm-3", 1, noble)
	newMM3 := w[0]

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, newMM3)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM3.FormattedAddress())
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

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Update Pauser while TF is paused -

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-1", 1, noble)
	newPauser1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser1.FormattedAddress())
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

	// - Update Pauser from non owner account -
	// Status:
	// 	Pauser: newPauser1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-2", 1, noble)
	newPauser2 := w[0]

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the owner: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-pauser")
	require.NoError(t, err, "failed to query show-pauser")

	err = json.Unmarshal(res, &getPauserResponse)
	require.NoError(t, err, "failed to unmarshall show-pauser response")

	require.Equal(t, expectedGetPauserResponse.Pauser, getPauserResponse.Pauser)

	// - Update Pauser from blacklisted owner account -
	// Status:
	// 	Pauser: newPauser1

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
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

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Owner)

	// - Update Pauser to blacklisted Pauser account -
	// Status:
	// 	Pauser: newPauser2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-3", 1, noble)
	newPauser3 := w[0]

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, newPauser3)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser3.FormattedAddress())
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

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Update Blacklister while TF is paused -

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-1", 1, noble)
	newBlacklister1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister1.FormattedAddress())
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

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Update Blacklister from non owner account -
	// Status:
	// 	Blacklister: newBlacklister1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-2", 1, noble)
	newBlacklister2 := w[0]

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the owner: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklister")
	require.NoError(t, err, "failed to query show-blacklister")

	err = json.Unmarshal(res, &getBlacklisterResponse)
	require.NoError(t, err, "failed to unmarshall show-blacklister response")

	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, getBlacklisterResponse.Blacklister)

	// - Update Blacklister from blacklisted owner account -
	// Status:
	// 	Blacklister: newBlacklister1

	blacklistAccount(t, ctx, val, newBlacklister1, gw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
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

	unblacklistAccount(t, ctx, val, newBlacklister2, gw.fiatTfRoles.Owner)

	// - Update Blacklister to blacklisted Blacklister account -
	// Status:
	// 	Blacklister: newBlacklister2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-3", 1, noble)
	newBlacklister3 := w[0]

	blacklistAccount(t, ctx, val, newBlacklister2, newBlacklister3)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister3.FormattedAddress())
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

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Blacklist user while TF is paused -

	w := interchaintest.GetAndFundTestUsers(t, ctx, "to-blacklist-1", 1, noble)
	toBlacklist1 := w[0]

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, toBlacklist1)

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Blacklist user from non Blacklister account -

	w = interchaintest.GetAndFundTestUsers(t, ctx, "to-blacklist-2", 1, noble)
	toBlacklist2 := w[0]

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	var preFailedBlacklist, postFailedBlacklist fiattokenfactorytypes.QueryAllBlacklistedResponse
	_ = json.Unmarshal(res, &preFailedBlacklist)
	// ignore the error since `pagination` does not unmarshall)

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the blacklister: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	_ = json.Unmarshal(res, &postFailedBlacklist)
	// ignore the error since `pagination` does not unmarshall)

	require.Equal(t, preFailedBlacklist.Blacklisted, postFailedBlacklist.Blacklisted)

	// Blacklist an account while the blacklister is blacklisted
	// Status:
	// 	blacklisted: toBlacklist1
	// 	not blacklisted: toBlacklist2

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Blacklister)

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, toBlacklist2)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Blacklister)

	// Blacklist an already blacklisted account
	// Status:
	// 	blacklisted: toBlacklist1, toBlacklist2

	hash, err = val.ExecTx(ctx, gw.fiatTfRoles.Blacklister.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "user is already blacklisted")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

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

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Unblacklist user while TF is paused -

	w := interchaintest.GetAndFundTestUsers(t, ctx, "blacklist-user-1", 1, noble)
	blacklistedUser1 := w[0]

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, blacklistedUser1)

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, blacklistedUser1)

	// - Unblacklist user from non Blacklister account -
	// Status:
	// 	not blacklisted: blacklistedUser1

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, blacklistedUser1)

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	var preFailedUnblacklist, postFailedUnblacklist fiattokenfactorytypes.QueryAllBlacklistedResponse
	_ = json.Unmarshal(res, &preFailedUnblacklist)
	// ignore the error since `pagination` does not unmarshall)

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "unblacklist", blacklistedUser1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the blacklister: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	_ = json.Unmarshal(res, &postFailedUnblacklist)
	// ignore the error since `pagination` does not unmarshall)

	require.Equal(t, preFailedUnblacklist.Blacklisted, postFailedUnblacklist.Blacklisted)

	// - Unblacklist an account while the blacklister is blacklisted -
	// Status:
	// 	not blacklisted: blacklistedUser1

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, blacklistedUser1)

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Blacklister)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, blacklistedUser1)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Blacklister)

	// - Unblacklist an account that is not blacklisted -
	// Status:
	// 	not blacklisted: blacklistedUser1

	hash, err = val.ExecTx(ctx, gw.fiatTfRoles.Blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", blacklistedUser1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "the specified address is not blacklisted")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", blacklistedUser1.FormattedAddress())
	require.Error(t, err, "query succeeded, blacklisted account should not exist")

}

func TestFiatTFPause(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Pause TF from an account that is not the Pauser -

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "pause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the pauser: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
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

	// - Pause TF from a blacklisted Pauser account
	// Status:
	// 	Paused: false

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Pauser)

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Pause TF while TF is already paused -
	// Status:
	// 	Paused: true

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

}

func TestFiatTFUnpause(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Unpause TF from an account that is not a Pauser

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "unpause")
	require.NoError(t, err, "error unpausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the pauser: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
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

	// - Unpause TF from a blacklisted Pauser account
	// Status:
	// 	Paused: true

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.Pauser)

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Pause TF while TF is already paused -
	// Status:
	// 	Paused: false

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)
}

func TestFiatTFConfigureMinterController(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Configure Minter Controller while TF is paused -

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", 1, noble)
	minterController1 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", 1, noble)
	minter1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
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

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Configure Minter Controller from non Master Minter account -
	// Status:
	// 	minterController1 -> minter1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-2", 1, noble)
	minterController2 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-2", 1, noble)
	minter2 := w[0]

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController2.FormattedAddress(), minter2.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the master minter: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController2.FormattedAddress())
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	// - Configure a blacklisted Minter Controller and Minter from blacklisted Master Minter account  -
	// Status:
	// 	minterController1 -> minter1

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.MasterMinter)
	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minterController2)
	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minter2)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController2.FormattedAddress(), minter2.FormattedAddress())
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

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.MasterMinter)

	// - Configure an already configured Minter Controller with a new Minter -
	// The old minter should be disascociated from Minter Controller but keep its status and allowance
	// Status:
	// 	minterController1 -> minter1
	// 	minterController2 -> minter2

	// confiuging minter to ensure allownace stays the same
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
				Amount: sdktypes.NewInt(1),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, getMintersPreUpdateMinterController.Minters, "configured minter and or allowance is not as expected")

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-3", 1, noble)
	minter3 := w[0]

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter3.FormattedAddress())
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

	// -- Configure an already configured Minter to another Minter Controller -
	// Status:
	// 	minterController1 -> minter3
	// 	minterController2 -> minter2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-3", 1, noble)
	minterController3 := w[0]

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController3.FormattedAddress(), minter3.FormattedAddress())
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

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Remove Minter Controller if TF is paused -

	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", 1, noble)
	minterController1 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", 1, noble)
	minter1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
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

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Remove a Minter Controller from non Master Minter account -

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")

	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)

	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "remove-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not the master minter: unauthorized")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter-controller")

	err = json.Unmarshal(res, &showMinterController)
	require.NoError(t, err)

	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMinterController.MinterController)

	// - Remove Minter Controller while Minter and Minter Controller are blacklisted
	// Status:
	// 	minterController1 -> minter1

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.MasterMinter)
	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minterController1)

	_, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minter-controller", minterController1.FormattedAddress())
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, gw.fiatTfRoles.MasterMinter)

	// - Remove a a non existent Minter Controller -

	hash, err = val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", minterController1.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, fmt.Sprintf("minter controller with a given address (%s) doesn't exist: user not found", minterController1.FormattedAddress()))
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

}

func TestFiatTFConfigureMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// @DAN: INVESTIGATE
	// - Configure minter while TF is paused -

	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", 1, noble)
	minterController1 := w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", 1, noble)
	minter1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
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

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	allowance := int64(10)

	hash, err := val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.NoError(t, err, "error configuring minter")
	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	fmt.Println("RESPONSE!!!", string(res))
	// _ = json.Unmarshal(res, &txResponse)
	// require.Contains(t, txResponse.RawLog, "minting is paused")
	// require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.Error(t, err, "minter found; configuring minter should not have succeeded")
	fmt.Println("RES!!", string(res))

	var showMinterResponse fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: sdktypes.NewInt(allowance),
			},
		},
	}

	// This should fail as these should not be be qual
	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

}

func TestFiatTFRemoveMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Remove minter while TF is paused -

	allowance := int64(10)

	minter1, minterController1 := setupMinter(t, ctx, noble, val, gw.fiatTfRoles.MasterMinter, allowance)

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	_, err := val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "remove-minter", minter1.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.Error(t, err, "minter found; not successfully removed")

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Remove minter from a minter controller not associated with the minter -
	// Status:
	// 	minterController1 -> minter1 (Removed)

	// reconfigure minter
	configureMinter(t, ctx, val, minterController1, minter1, allowance)

	_, minterController2 := setupMinter(t, ctx, noble, val, gw.fiatTfRoles.MasterMinter, allowance)

	hash, err := val.ExecTx(ctx, minterController2.KeyName(), "fiat-tokenfactory", "remove-minter", minter1.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "minter address ≠ minter controller's minter address")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	// ensure minter still exists
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")

	var showMinterResponse fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: sdktypes.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	// - Remove minter from a blacklisted minter controller -
	// Status:
	// 	minterController1 -> minter1
	// 	minterController2 -> minter2

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minterController1)

	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "remove-minter", minter1.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.Error(t, err, "minter found; not successfully removed")

}

func TestFiatTFMint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// - Mint while TF is paused -

	allowance := int64(10)
	minter1, _ := setupMinter(t, ctx, noble, val, gw.fiatTfRoles.MasterMinter, allowance)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "receiver-1", 1, noble)
	receiver1 := w[0]

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	hash, err := val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error minting")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "minting is paused")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err := noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")

	require.Zero(t, bal)

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")

	var showMinterResponse fiattokenfactorytypes.QueryGetMintersResponse
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: sdktypes.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Mint from non minter -

	hash, err = val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error minting")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not a minter")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")

	require.Zero(t, bal)

	// - Mint from blacklisted minter -

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minter1)

	hash, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error minting")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "minter address is blacklisted")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")

	require.Zero(t, bal)

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minter1)

	// - Mint to blacklisted account -

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, receiver1)

	hash, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error minting")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "receiver address is blacklisted")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")

	require.Zero(t, bal)

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, receiver1)

	// - Mint an amount that exceeds the minters allowance -

	hash, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "100uusdc")
	require.NoError(t, err, "error minting")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "minting amount is greater than the allowance")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")

	require.Zero(t, bal)

	// allowance should not have changed
	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

	// - Successfully mint into an account -

	mintAmount := int64(3)
	_, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")

	require.Equal(t, mintAmount, bal)

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-minters", minter1.FormattedAddress())
	require.NoError(t, err, "failed to query show-minter")
	err = json.Unmarshal(res, &showMinterResponse)
	require.NoError(t, err, "failed to unmarshall")

	expectedShowMinters.Minters.Allowance = sdktypes.Coin{
		Denom:  "uusdc",
		Amount: sdktypes.NewInt(allowance - mintAmount),
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)

}

func TestFiatTFBurn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(t, ctx)
	noble := gw.chain
	val := noble.Validators[0]

	// setup
	allowance := int64(15)
	minter1, _ := setupMinter(t, ctx, noble, val, gw.fiatTfRoles.MasterMinter, allowance)

	mintAmount := int64(10)
	_, err := val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "mint", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err := noble.GetBalance(ctx, minter1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, mintAmount, bal)

	// - Burn while TF is paused -

	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	burnAmount := int64(3)
	hash, err := val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "burning is paused")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, minter1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, mintAmount, bal)

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser)

	// - Burn from non minter account -

	hash, err = val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "you are not a minter")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, minter1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, mintAmount, bal)

	// - Burn from a blacklisted minter account -

	blacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minter1)

	hash, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "minter address is blacklisted")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, minter1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, mintAmount, bal)

	unblacklistAccount(t, ctx, val, gw.fiatTfRoles.Blacklister, minter1)

	// - Burn amount greater than the minter allowance -

	hash, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "burn", "999uusdc")
	require.NoError(t, err, "error broadcasting burn")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")
	_ = json.Unmarshal(res, &txResponse)
	require.Contains(t, txResponse.RawLog, "insufficient funds")
	require.Greater(t, txResponse.Code, uint32(0), "got 'successful' code response")

	bal, err = noble.GetBalance(ctx, minter1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, mintAmount, bal)

	// - Burn succeeds -

	_, err = val.ExecTx(ctx, minter1.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	bal, err = noble.GetBalance(ctx, minter1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.Equal(t, mintAmount-burnAmount, bal)

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

// setupMinter creates a minter and minter controller. It also sets up a minter with an specified allowance of `uusdc`
func setupMinter(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, val *cosmos.ChainNode, masterMinter ibc.Wallet, allowance int64) (minter ibc.Wallet, minterController ibc.Wallet) {
	w := interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-1", 1, noble)
	minterController = w[0]
	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-1", 1, noble)
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

// configureMinter configures a minter with a specified allowance of `uusdc`
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
				Amount: sdktypes.NewInt(allowance),
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterResponse.Minters)
}

// nobleSpinUp starts noble chain and sets up Fiat Token Factory Roles
func nobleSpinUp(t *testing.T, ctx context.Context) (gw genesisWrapper) {
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", nv, nf, true, false, true, false),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain

	// cmd.SetPrefixes(noble.Config().Bech32Prefix)

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
