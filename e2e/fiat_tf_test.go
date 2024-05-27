package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"
)

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

	showMMRes, err := showMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse := fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: string(newMM1.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Update Master Minter from non owner account
	// EXPECTED: Request fails; Master Minter not updated
	// Status:
	// 	Master Minter: newMM1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newMM2 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	showMMRes, err = showMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

	// ACTION: Update Master Minter from blacklisted owner account
	// EXPECTED: Success; Master Minter updated
	// Status:
	// 	Master Minter: newMM1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	showMMRes, err = showMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: string(newMM2.FormattedAddress()),
		},
	}

	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

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

	showMMRes, err = showMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: string(newMM3.FormattedAddress()),
		},
	}
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)
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

	showMCRes, err := showMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Configure Minter Controller from non Master Minter account
	// EXPECTED: Request fails; Minter Controller not configured with Minter
	// Status:
	// 	minterController1 -> minter1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble, noble)
	minterController2 := w[0]
	minter2 := w[1]
	alice := w[2]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController2.FormattedAddress(), minter2.FormattedAddress())
	require.ErrorContains(t, err, "you are not the master minter: unauthorized")

	_, err = showMinterController(ctx, val, minterController2)
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

	showMCRes, err = showMinterController(ctx, val, minterController2)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter2.FormattedAddress(),
			Controller: minterController2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)

	// ACTION: Configure an already configured Minter Controller with a new Minter
	// EXPECTED: Success; Minter Controller is configured with Minter. The old minter should be disassociated
	// from Minter Controller but keep its status and allowance
	// Status:
	// 	minterController1 -> minter1
	// 	minterController2 -> minter2

	// configuring minter1 to ensure allowance stays the same after assigning minterController1 a new minter
	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error configuring minter controller")

	showMinterPreUpdateMinterController, err := showMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.OneInt(),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, showMinterPreUpdateMinterController.Minters, "configured minter and or allowance is not as expected")

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-3", math.OneInt(), noble)
	minter3 := w[0]

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter3.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = showMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController, "expected minter and minter controller is not as expected")

	showMinterPostUpdateMinterController, err := showMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minter")
	require.Equal(t, showMinterPreUpdateMinterController.Minters, showMinterPostUpdateMinterController.Minters, "the minter should not have changed since updating the minter controller with a new minter")

	// ACTION:- Configure an already configured Minter to another Minter Controller
	// EXPECTED: Success; Minter Controller is configured with new Minter. Minter can have multiple Minter Controllers.
	// Status:
	// 	minterController1 -> minter3
	// 	minterController2 -> minter2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-3", math.OneInt(), noble)
	minterController3 := w[0]

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController3.FormattedAddress(), minter3.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = showMinterController(ctx, val, minterController3)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController3.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	showMCRes, err = showMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-minter-controller")
	require.NoError(t, err, "failed to query list-minter-controller")
	var listMinterController fiattokenfactorytypes.QueryAllMinterControllerResponse
	_ = json.Unmarshal(res, &listMinterController)
	// ignore error because `pagination` does not unmarshal

	expectedListMinterController := fiattokenfactorytypes.QueryAllMinterControllerResponse{
		MinterController: []fiattokenfactorytypes.MinterController{
			{ // this minter and controller were created/assigned at genesis
				Minter:     nw.fiatTfRoles.Minter.FormattedAddress(),
				Controller: nw.fiatTfRoles.MinterController.FormattedAddress(),
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

	_, err = showMinterController(ctx, val, nw.fiatTfRoles.MinterController)
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	unpauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	// ACTION: Remove a Minter Controller from non Master Minter account
	// EXPECTED: Request fails; Minter Controller remains configured with Minter

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress(), nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err := showMinterController(ctx, val, nw.fiatTfRoles.MinterController)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     nw.fiatTfRoles.Minter.FormattedAddress(),
			Controller: nw.fiatTfRoles.MinterController.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.ErrorContains(t, err, "you are not the master minter: unauthorized")

	showMCRes, err = showMinterController(ctx, val, nw.fiatTfRoles.MinterController)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     nw.fiatTfRoles.Minter.FormattedAddress(),
			Controller: nw.fiatTfRoles.MinterController.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	// ACTION: Remove Minter Controller while Minter and Minter Controller are blacklisted
	// EXPECTED: Success; Minter Controller is removed
	// Status:
	// 	gw minterController -> gw minter

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)
	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MinterController)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.fiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, err = showMinterController(ctx, val, nw.fiatTfRoles.MinterController)
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	unblacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MasterMinter)

	// ACTION: Remove a a non existent Minter Controller
	// EXPECTED: Request fails
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
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	minterController1 := w[0]
	minter1 := w[1]

	_, err := val.ExecTx(ctx, nw.fiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err := showMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	pauseFiatTF(t, ctx, val, nw.fiatTfRoles.Pauser)

	allowance := int64(10)

	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.ErrorContains(t, err, "minting is paused")

	_, err = showMinters(ctx, val, minter1)
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

	showMintersRes, err := showMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minters")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters, "configured minter allowance is not as expected")

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

	_, err = showMinters(ctx, val, nw.fiatTfRoles.Minter)
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
	showMintersRes, err := showMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdktypes.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	// ACTION: Remove minter from a blacklisted minter controller
	// EXPECTED: Success; Minter is removed
	// Status:
	// 	gw minterController -> gw minter
	// 	minterController1 -> minter1

	blacklistAccount(t, ctx, val, nw.fiatTfRoles.Blacklister, nw.fiatTfRoles.MinterController)

	_, err = val.ExecTx(ctx, nw.fiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", nw.fiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, err = showMinters(ctx, val, nw.fiatTfRoles.Minter)
	require.Error(t, err, "minter found; not successfully removed")
}
