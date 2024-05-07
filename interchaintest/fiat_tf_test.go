package interchaintest_test

import (
	"context"
	"encoding/json"
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

func TestFiatTFUpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(ctx, t)
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

	// TODO: after updating to latest interchaintest, replaces all these tx hash queries with: noble.GetTransaction(hash)
	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")

	var txResponse sdktypes.TxResponse
	_ = json.Unmarshal(res, &txResponse)
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")
}

func TestFiatTFAcceptOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(ctx, t)
	noble := gw.chain
	val := noble.Validators[0]

	// - Accept owner while TF is paused -

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", 1, noble)
	newOwner1 := w[0]

	hash, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", newOwner1.FormattedAddress())
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")

	var txResponse sdktypes.TxResponse
	_ = json.Unmarshal(res, &txResponse)
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

	require.Equal(t, uint32(0), txResponse.Code, "update owner failed")

	hash, err = val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "accept-owner")
	require.NoError(t, err, "failed to accept owner")

	res, _, err = val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")

	_ = json.Unmarshal(res, &txResponse)
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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

	gw := nobleSpinUp(ctx, t)
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

	var txResponse sdktypes.TxResponse
	_ = json.Unmarshal(res, &txResponse)
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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

// wip
func TestFiatTFUpdatePauser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(ctx, t)
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

	var txResponse sdktypes.TxResponse
	_ = json.Unmarshal(res, &txResponse)
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

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

func unblacklistAccount(t *testing.T, ctx context.Context, val *cosmos.ChainNode, blacklister ibc.Wallet, unBlacklist ibc.Wallet) {
	_, err := val.ExecTx(ctx, blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", unBlacklist.FormattedAddress())
	require.NoError(t, err, "failed to broadcast blacklist message")

	_, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "show-blacklisted", unBlacklist.FormattedAddress())
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
}

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

// starts noble chain and sets up all Fiat Token Factory Roles
func nobleSpinUp(ctx context.Context, t *testing.T) (gw genesisWrapper) {
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
