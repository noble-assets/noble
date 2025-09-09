// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
)

func TestFiatTFUpdateOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
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
	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
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

func TestFiatTFUpdateMasterMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy Path: Update Master Minter
	// EXPECTED: Success; Master Minter updated

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-masterMinter-1", math.OneInt(), noble)
	newMM1 := w[0]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	showMMRes, err := e2e.ShowMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse := fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: newMM1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

	// ACTION: Update Master Minter while TF is paused
	// EXPECTED: Success; Master Minter updated
	// Status:
	// 	Master Minter: newMM1

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-masterMinter-2", math.OneInt(), noble)
	newMM2 := w[0]

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	showMMRes, err = e2e.ShowMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: newMM2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Update Master Minter from non owner account
	// EXPECTED: Request fails; Master Minter not updated
	// Status:
	// 	Master Minter: newMM2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newMM3 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM3.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	showMMRes, err = e2e.ShowMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

	// ACTION: Update Master Minter from blacklisted owner account
	// EXPECTED: Success; Master Minter updated
	// Status:
	// 	Master Minter: newMM2

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM3.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	showMMRes, err = e2e.ShowMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: newMM3.FormattedAddress(),
		},
	}

	require.Equal(t, expectedGetMasterMinterResponse.MasterMinter, showMMRes.MasterMinter)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Owner)

	// ACTION: Update Master Minter to blacklisted Master Minter account
	// EXPECTED: Success; Master Minter updated
	// Status:
	// 	Master Minter: newMM3

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-mm-4", math.OneInt(), noble)
	newMM4 := w[0]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, newMM4)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-master-minter", newMM4.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-master-minter message")

	showMMRes, err = e2e.ShowMasterMinter(ctx, val)
	require.NoError(t, err, "failed to query show-master-minter")
	expectedGetMasterMinterResponse = fiattokenfactorytypes.QueryGetMasterMinterResponse{
		MasterMinter: fiattokenfactorytypes.MasterMinter{
			Address: newMM4.FormattedAddress(),
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

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Configure Minter Controller
	// EXPECTED: Success; Minter Controller is configured with Minter

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	minterController1 := w[0]
	minter1 := w[1]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err := e2e.ShowMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	// ACTION: Configure Minter Controller while TF is paused
	// EXPECTED: Success; Minter Controller is configured with Minter
	// Status:
	// 	minterController1 -> minter1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	minterController2 := w[0]
	minter2 := w[1]

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController2.FormattedAddress(), minter2.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = e2e.ShowMinterController(ctx, val, minterController2)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter2.FormattedAddress(),
			Controller: minterController2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Configure Minter Controller from non Master Minter account
	// EXPECTED: Request fails; Minter Controller not configured with Minter
	// Status:
	// 	minterController1 -> minter1
	// 	minterController2 -> minter2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble, noble)
	minterController3 := w[0]
	minter3 := w[1]
	alice := w[2]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController3.FormattedAddress(), minter3.FormattedAddress())
	require.ErrorContains(t, err, "you are not the master minter: unauthorized")

	_, err = e2e.ShowMinterController(ctx, val, minterController3)
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	// ACTION: Configure a blacklisted Minter Controller and Minter from blacklisted Master Minter account
	// EXPECTED: Success; Minter Controller is configured with Minter
	// Status:
	//  minterController1 -> minter1
	// 	minterController2 -> minter2

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.MasterMinter)
	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, minterController2)
	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, minter2)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController3.FormattedAddress(), minter3.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = e2e.ShowMinterController(ctx, val, minterController3)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter3.FormattedAddress(),
			Controller: minterController3.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.MasterMinter)

	// ACTION: Configure an already configured Minter Controller with a new Minter
	// EXPECTED: Success; Minter Controller is configured with Minter. The old minter should be disassociated
	// from Minter Controller but keep its status and allowance
	// Status:
	// 	minterController1 -> minter1
	// 	minterController2 -> minter2
	//  minterController3 -> minter3

	// configuring minter1 to ensure allowance stays the same after assigning minterController1 a new minter
	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), "1uusdc")
	require.NoError(t, err, "error configuring minter controller")

	showMinterPreUpdateMinterController, err := e2e.ShowMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdk.Coin{
				Denom:  "uusdc",
				Amount: math.OneInt(),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, showMinterPreUpdateMinterController.Minters, "configured minter and or allowance is not as expected")

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-4", math.OneInt(), noble)
	minter4 := w[0]

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter4.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = e2e.ShowMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter4.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController, "expected minter and minter controller is not as expected")

	showMinterPostUpdateMinterController, err := e2e.ShowMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minter")
	require.Equal(t, showMinterPreUpdateMinterController.Minters, showMinterPostUpdateMinterController.Minters, "the minter should not have changed since updating the minter controller with a new minter")

	// ACTION:- Configure an already configured Minter to another Minter Controller
	// EXPECTED: Success; Minter Controller is configured with new Minter. Minter can have multiple Minter Controllers.
	// Status:
	// 	minterController1 -> minter4
	// 	minterController2 -> minter2
	// 	minterController3 -> minter3
	//
	//  minter1 has a minting allowance but is not controlled by any minterController

	w = interchaintest.GetAndFundTestUsers(t, ctx, "minter-controller-4", math.OneInt(), noble)
	minterController4 := w[0]

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController4.FormattedAddress(), minter4.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = e2e.ShowMinterController(ctx, val, minterController4)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter4.FormattedAddress(),
			Controller: minterController4.FormattedAddress(),
		},
	}
	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	showMCRes, err = e2e.ShowMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController = fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter4.FormattedAddress(),
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
				Minter:     nw.FiatTfRoles.Minter.FormattedAddress(),
				Controller: nw.FiatTfRoles.MinterController.FormattedAddress(),
			},
			{
				Minter:     minter4.FormattedAddress(),
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
			{
				Minter:     minter4.FormattedAddress(),
				Controller: minterController4.FormattedAddress(),
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

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Remove Minter Controller
	// EXPECTED: Success; Minter Controller is removed

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, err = e2e.ShowMinterController(ctx, val, nw.FiatTfRoles.MinterController)
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	// ACTION: Remove Minter Controller while TF is paused
	// EXPECTED: Success; Minter Controller is removed

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress(), nw.FiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err := e2e.ShowMinterController(ctx, val, nw.FiatTfRoles.MinterController)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     nw.FiatTfRoles.Minter.FormattedAddress(),
			Controller: nw.FiatTfRoles.MinterController.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, err = e2e.ShowMinterController(ctx, val, nw.FiatTfRoles.MinterController)
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Remove a Minter Controller from non Master Minter account
	// EXPECTED: Request fails; Minter Controller remains configured with Minter

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress(), nw.FiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err = e2e.ShowMinterController(ctx, val, nw.FiatTfRoles.MinterController)
	require.NoError(t, err, "failed to query show-minter-controller")

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress())
	require.ErrorContains(t, err, "you are not the master minter: unauthorized")

	showMCRes, err = e2e.ShowMinterController(ctx, val, nw.FiatTfRoles.MinterController)
	require.NoError(t, err, "failed to query show-minter-controller")

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	// ACTION: Remove Minter Controller while MasterMinter and Minter Controller are blacklisted
	// EXPECTED: Success; Minter Controller is removed
	// Status:
	// 	gw minterController -> gw minter

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.MasterMinter)
	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.MinterController)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress())
	require.NoError(t, err, "error removing minter controller")

	_, err = e2e.ShowMinterController(ctx, val, nw.FiatTfRoles.MinterController)
	require.Error(t, err, "successfully queried for the minter controller when it should have failed")

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.MasterMinter)

	// ACTION: Remove a a non existent Minter Controller
	// EXPECTED: Request fails
	// Status:
	// 	no minterController setup

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "remove-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress())
	require.ErrorContains(t, err, fmt.Sprintf("minter controller with a given address (%s) doesn't exist: user not found", nw.FiatTfRoles.MinterController.FormattedAddress()))
}

func TestFiatTFConfigureMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Configure minter
	// EXPECTED: Success; Minter is configured with allowance

	e2e.ConfigureMinter(t, ctx, val, nw.FiatTfRoles.MinterController, nw.FiatTfRoles.Minter, 20)

	// ACTION: Configure minter while TF is paused
	// EXPECTED: Request fails; Minter is not configured

	// configure new minter controller and minter
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	minterController1 := w[0]
	minter1 := w[1]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(), "fiat-tokenfactory", "configure-minter-controller", minterController1.FormattedAddress(), minter1.FormattedAddress())
	require.NoError(t, err, "error configuring minter controller")

	showMCRes, err := e2e.ShowMinterController(ctx, val, minterController1)
	require.NoError(t, err, "failed to query show-minter-controller")
	expectedShowMinterController := fiattokenfactorytypes.QueryGetMinterControllerResponse{
		MinterController: fiattokenfactorytypes.MinterController{
			Minter:     minter1.FormattedAddress(),
			Controller: minterController1.FormattedAddress(),
		},
	}

	require.Equal(t, expectedShowMinterController.MinterController, showMCRes.MinterController)

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	allowance := int64(10)

	_, err = val.ExecTx(ctx, minterController1.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", allowance))
	require.ErrorContains(t, err, "minting is paused")

	_, err = e2e.ShowMinters(ctx, val, minter1)
	require.Error(t, err, "minter found; configuring minter should not have succeeded")

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Configure minter from a minter controller not associated with the minter
	// EXPECTED: Request fails; Minter is not configured with new Minter Controller but old Minter retains its allowance
	// Status:
	// 	minterController1 -> minter1 (un-configured)
	// 	gw minterController -> gw minter

	// reconfigure minter1 to ensure balance does not change
	e2e.ConfigureMinter(t, ctx, val, minterController1, minter1, allowance)

	// reconfigure minter1 with a new allowance from a minter controller not associated with the minter
	differentAllowance := allowance + 99
	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "configure-minter", minter1.FormattedAddress(), fmt.Sprintf("%duusdc", differentAllowance))
	require.ErrorContains(t, err, "minter address ≠ minter controller's minter address")

	showMintersRes, err := e2e.ShowMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minters")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdk.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters, "configured minter allowance is not as expected")

	// ACTION: Configure a minter is blacklisted from a blacklisted minter controller
	// EXPECTED: Success; Minter is configured with allowance
	// Status:
	// minterController1 -> minter1
	// gw minterController -> gw minter

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, minterController1)

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, minter1)

	e2e.ConfigureMinter(t, ctx, val, minterController1, minter1, 11)
}

func TestFiatTFRemoveMinter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Remove minter
	// EXPECTED: Success; Minter is removed

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", nw.FiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, err = e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.Error(t, err, "minter found; not successfully removed")

	// ACTION: Remove minter while TF is paused
	// EXPECTED: Success; Minter is removed

	allowance := int64(10)

	// reconfigure minter
	e2e.ConfigureMinter(t, ctx, val, nw.FiatTfRoles.MinterController, nw.FiatTfRoles.Minter, allowance)

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", nw.FiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, err = e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.Error(t, err, "minter found; not successfully removed")

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Remove minter from a minter controller not associated with the minter
	// EXPECTED: Request fails; Minter is not removed
	// Status:
	// 	gw minterController -> gw minter (Removed)

	// reconfigure minter
	e2e.ConfigureMinter(t, ctx, val, nw.FiatTfRoles.MinterController, nw.FiatTfRoles.Minter, allowance)

	minter1, _ := e2e.SetupMinterAndController(t, ctx, noble, val, nw.FiatTfRoles.MasterMinter, allowance)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", minter1.FormattedAddress())
	require.ErrorContains(t, err, "minter address ≠ minter controller's minter address")

	// ensure minter still exists
	showMintersRes, err := e2e.ShowMinters(ctx, val, minter1)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: minter1.FormattedAddress(),
			Allowance: sdk.Coin{
				Denom:  "uusdc",
				Amount: math.NewInt(allowance),
			},
		},
	}
	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	// ACTION: Remove blacklisted minter from a blacklisted minter controller
	// EXPECTED: Success; Minter is removed
	// Status:
	// 	gw minterController -> gw minter
	// 	minterController1 -> minter1

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.MinterController)

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(), "fiat-tokenfactory", "remove-minter", nw.FiatTfRoles.Minter.FormattedAddress())
	require.NoError(t, err, "error broadcasting removing minter")

	_, err = e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.Error(t, err, "minter found; not successfully removed")
}

func TestFiatTFUpdatePauser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Update Pauser
	// EXPECTED: Success; pauser updated

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-1", math.OneInt(), noble)
	newPauser1 := w[0]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err := e2e.ShowPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse := fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: newPauser1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	// ACTION: Update Pauser while TF is paused
	// EXPECTED: Success; pauser updated
	// Status:
	// 	Pauser: newPauser1

	e2e.PauseFiatTF(t, ctx, val, newPauser1)

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-2", math.OneInt(), noble)
	newPauser2 := w[0]

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err = e2e.ShowPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: newPauser2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	e2e.UnpauseFiatTF(t, ctx, val, newPauser2)

	// ACTION: Update Pauser from non owner account
	// EXPECTED: Request fails; pauser not updated
	// Status:
	// 	Pauser: newPauser2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newPauser3 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser3.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	showPauserRes, err = e2e.ShowPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	// ACTION: Update Pauser from blacklisted owner account
	// EXPECTED: Success; pauser updated
	// Status:
	// 	Pauser: newPauser2

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser3.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err = e2e.ShowPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: newPauser3.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetPauserResponse.Pauser, showPauserRes.Pauser)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Owner)

	// ACTION: Update Pauser to blacklisted Pauser account
	// EXPECTED: Success; pauser updated
	// Status:
	// 	Pauser: newPauser3

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-pauser-4", math.OneInt(), noble)
	newPauser4 := w[0]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, newPauser4)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-pauser", newPauser4.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-pauser message")

	showPauserRes, err = e2e.ShowPauser(ctx, val)
	require.NoError(t, err, "failed to query show-pauser")
	expectedGetPauserResponse = fiattokenfactorytypes.QueryGetPauserResponse{
		Pauser: fiattokenfactorytypes.Pauser{
			Address: newPauser4.FormattedAddress(),
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

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Pause TF
	// EXPECTED: Success; TF is paused

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Pause TF from an account that is not the Pauser
	// EXPECTED: Request fails; TF not paused
	// Status:
	// 	Paused: true

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	_, err := val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "pause")
	require.ErrorContains(t, err, "you are not the pauser: unauthorized")

	showPausedRes, err := e2e.ShowPaused(ctx, val)
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

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Pauser)

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Pauser)

	// ACTION: Pause TF while TF is already paused
	// EXPECTED: Success; TF remains paused
	// Status:
	// 	Paused: true

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)
}

func TestFiatTFUnpause(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Unpause TF
	// EXPECTED: Success; TF is unpaused

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Unpause TF from an account that is not a Pauser
	// EXPECTED: Request fails; TF remains paused

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err := val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "unpause")
	require.ErrorContains(t, err, "you are not the pauser: unauthorized")

	showPausedRes, err := e2e.ShowPaused(ctx, val)
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

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Pauser)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Pauser)

	// ACTION: Unpause TF while TF is already unpaused
	// EXPECTED: Success; TF remains unpaused
	// Status:
	// 	Paused: false

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)
}

func TestFiatTFUpdateBlacklister(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Update Blacklister
	// EXPECTED: Success; blacklister updated

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-1", math.OneInt(), noble)
	newBlacklister1 := w[0]

	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister1.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err := e2e.ShowBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse := fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: newBlacklister1.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	// ACTION: Update Blacklister while TF is paused
	// EXPECTED: Success; blacklister updated
	// Status:
	// 	Blacklister: newBlacklister1

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-2", math.OneInt(), noble)
	newBlacklister2 := w[0]

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister2.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err = e2e.ShowBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: newBlacklister2.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Update Blacklister from non owner account
	// EXPECTED: Request fails; blacklister not updated
	// Status:
	// 	Blacklister: newBlacklister2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	newBlacklister3 := w[0]
	alice := w[1]

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister3.FormattedAddress())
	require.ErrorContains(t, err, "you are not the owner: unauthorized")

	showBlacklisterRes, err = e2e.ShowBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	// ACTION: Update Blacklister from blacklisted owner account
	// EXPECTED: Success; blacklister updated
	// Status:
	// 	Blacklister: newBlacklister2

	e2e.BlacklistAccount(t, ctx, val, newBlacklister2, nw.FiatTfRoles.Owner)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister3.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err = e2e.ShowBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: newBlacklister3.FormattedAddress(),
		},
	}
	require.Equal(t, expectedGetBlacklisterResponse.Blacklister, showBlacklisterRes.Blacklister)

	e2e.UnblacklistAccount(t, ctx, val, newBlacklister3, nw.FiatTfRoles.Owner)

	// ACTION: Update Blacklister to blacklisted Blacklister account
	// EXPECTED: Success; blacklister updated
	// Status:
	// 	Blacklister: newBlacklister3

	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-blacklister-4", math.OneInt(), noble)
	newBlacklister4 := w[0]

	e2e.BlacklistAccount(t, ctx, val, newBlacklister3, newBlacklister4)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-blacklister", newBlacklister4.FormattedAddress())
	require.NoError(t, err, "failed to broadcast update-blacklister message")

	showBlacklisterRes, err = e2e.ShowBlacklister(ctx, val)
	require.NoError(t, err, "failed to query show-blacklister")
	expectedGetBlacklisterResponse = fiattokenfactorytypes.QueryGetBlacklisterResponse{
		Blacklister: fiattokenfactorytypes.Blacklister{
			Address: newBlacklister4.FormattedAddress(),
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

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Blacklist user
	// EXPECTED: Success; user blacklisted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "to-blacklist-1", math.OneInt(), noble)
	toBlacklist1 := w[0]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, toBlacklist1)

	// ACTION: Blacklist user while TF is paused
	// EXPECTED: Success; user blacklisted
	// Status:
	// 	blacklisted: toBlacklist1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "to-blacklist-2", math.OneInt(), noble)
	toBlacklist2 := w[0]

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, toBlacklist2)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Blacklist user from non Blacklister account
	// EXPECTED: Request failed; user not blacklisted
	// Status:
	// 	blacklisted: toBlacklist1, toBlacklist2

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	toBlacklist3 := w[0]
	alice := w[1]

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")

	var preFailedBlacklist, postFailedBlacklist fiattokenfactorytypes.QueryAllBlacklistedResponse
	_ = json.Unmarshal(res, &preFailedBlacklist)
	// ignore the error since `pagination` does not unmarshal)
	require.NotContains(t, preFailedBlacklist.Blacklisted, toBlacklist3.Address())

	_, err = val.ExecTx(ctx, alice.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist3.FormattedAddress())
	require.ErrorContains(t, err, "you are not the blacklister: unauthorized")

	res, _, err = val.ExecQuery(ctx, "fiat-tokenfactory", "list-blacklisted")
	require.NoError(t, err, "failed to query list-blacklisted")
	_ = json.Unmarshal(res, &postFailedBlacklist)
	// ignore the error since `pagination` does not unmarshal)
	require.ElementsMatch(t, preFailedBlacklist.Blacklisted, postFailedBlacklist.Blacklisted)

	// Blacklist an account while the blacklister is blacklisted
	// EXPECTED: Success; user blacklisted
	// Status:
	// 	blacklisted: toBlacklist1, toBlacklist2
	// 	not blacklisted: toBlacklist3

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Blacklister)

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, toBlacklist3)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Blacklister)

	// Blacklist an already blacklisted account
	// EXPECTED: Request fails; user remains blacklisted
	// Status:
	// 	blacklisted: toBlacklist1, toBlacklist2, toBlacklist3

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Blacklister.KeyName(), "fiat-tokenfactory", "blacklist", toBlacklist1.FormattedAddress())
	require.ErrorContains(t, err, "user is already blacklisted")

	showBlacklistedRes, err := e2e.ShowBlacklisted(ctx, val, toBlacklist1)
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

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Happy path: Unblacklist user
	// EXPECTED: Success; user unblacklisted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "blacklist-user-1", math.OneInt(), noble)
	blacklistedUser1 := w[0]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, blacklistedUser1)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, blacklistedUser1)

	// ACTION: Unblacklist user while TF is paused
	// EXPECTED: Success; user unblacklisted
	// Status:
	// 	not blacklisted: blacklistedUser1

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, blacklistedUser1)

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, blacklistedUser1)

	// ACTION: Unblacklist user from non Blacklister account
	// EXPECTED: Request fails; user not unblacklisted
	// Status:
	// 	not blacklisted: blacklistedUser1

	w = interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.OneInt(), noble)
	alice := w[0]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, blacklistedUser1)

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

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Blacklister)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, blacklistedUser1)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Blacklister)

	// ACTION: Unblacklist an account that is not blacklisted
	// EXPECTED: Request fails; user remains unblacklisted
	// Status:
	// 	not blacklisted: blacklistedUser1

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Blacklister.KeyName(), "fiat-tokenfactory", "unblacklist", blacklistedUser1.FormattedAddress())
	require.ErrorContains(t, err, "the specified address is not blacklisted")

	_, err = e2e.ShowBlacklisted(ctx, val, blacklistedUser1)
	require.Error(t, err, "query succeeded, blacklisted account should not exist")
}

func TestFiatTFMint(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Mint while TF is paused
	// EXPECTED: Request fails; amount not minted

	w := interchaintest.GetAndFundTestUsers(t, ctx, "receiver-1", math.OneInt(), noble)
	receiver1 := w[0]

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	showMinterPreMint, err := e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")

	preMintAllowance := showMinterPreMint.Minters.Allowance.Amount

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "minting is paused")

	bal, err := noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMinterPostMint, err := e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters := fiattokenfactorytypes.QueryGetMintersResponse{
		Minters: fiattokenfactorytypes.Minters{
			Address: nw.FiatTfRoles.Minter.FormattedAddress(),
			Allowance: sdk.Coin{
				Denom:  "uusdc",
				Amount: preMintAllowance,
			},
		},
	}

	require.Equal(t, expectedShowMinters.Minters, showMinterPostMint.Minters)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

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

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "minter address is blacklisted")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMintersRes, err := e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")

	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Minter)

	// ACTION: Mint to blacklisted account
	// EXPECTED: Request fails; amount not minted

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, receiver1)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), "1uusdc")
	require.ErrorContains(t, err, "receiver address is blacklisted")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMintersRes, err = e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")

	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, receiver1)

	// ACTION: Mint an amount that exceeds the minters allowance
	// EXPECTED: Request fails; amount not minted

	exceedAllowance := preMintAllowance.Add(math.NewInt(99))
	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", exceedAllowance.Int64()))
	require.ErrorContains(t, err, "minting amount is greater than the allowance")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.IsZero())

	// allowance should not have changed
	showMintersRes, err = e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")
	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)

	// ACTION: Successfully mint into an account
	// EXPECTED: Success

	mintAmount := int64(3)
	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", receiver1.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err = noble.GetBalance(ctx, receiver1.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bal.Equal(math.NewInt(mintAmount)))

	showMintersRes, err = e2e.ShowMinters(ctx, val, nw.FiatTfRoles.Minter)
	require.NoError(t, err, "failed to query show-minter")
	expectedShowMinters.Minters.Allowance = sdk.Coin{
		Denom:  "uusdc",
		Amount: preMintAllowance.Sub(math.NewInt(mintAmount)),
	}

	require.Equal(t, expectedShowMinters.Minters, showMintersRes.Minters)
}

func TestFiatTFBurn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// setup - mint into minter's wallet
	mintAmount := int64(5)
	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nw.FiatTfRoles.Minter.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	bal, err := noble.GetBalance(ctx, nw.FiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64())

	// ACTION: Burn while TF is paused
	// EXPECTED: Request fails; amount not burned

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	burnAmount := int64(1)
	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "burning is paused")

	bal, err = noble.GetBalance(ctx, nw.FiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Burn from non minter account
	// EXPECTED: Request fails; amount not burned

	w := interchaintest.GetAndFundTestUsers(t, ctx, "alice", math.NewInt(burnAmount), noble)
	alice := w[0]

	// mint into Alice's account to give her a balance to burn
	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
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

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Minter)

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.ErrorContains(t, err, "minter address is blacklisted")

	bal, err = noble.GetBalance(ctx, nw.FiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nw.FiatTfRoles.Minter)

	// ACTION: Burn amount greater than the minters balance
	// EXPECTED: Request fails; amount not burned

	exceedAllowance := bal.Add(math.NewInt(99))
	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", exceedAllowance.Int64()))
	require.ErrorContains(t, err, "insufficient funds")

	bal, err = noble.GetBalance(ctx, nw.FiatTfRoles.Minter.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, mintAmount, bal.Int64(), "minters balance should not have decreased")

	// ACTION: Successfully burn tokens
	// EXPECTED: Success; amount burned and Minters balance is decreased

	_, err = val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "burn", fmt.Sprintf("%duusdc", burnAmount))
	require.NoError(t, err, "error broadcasting burn")

	bal, err = noble.GetBalance(ctx, nw.FiatTfRoles.Minter.FormattedAddress(), "uusdc")
	expectedAmount := mintAmount - burnAmount
	require.NoError(t, err, "error getting balance")
	require.EqualValues(t, expectedAmount, bal.Int64())
}

func TestFiatTFAuth(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Send non TF token but pay fee in TF token while the TF is paused
	// EXPECTED: Request fails; TF is paused

	originalAmount := math.OneInt()
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", originalAmount, noble, noble)
	alice := w[0] // 1ustake
	bob := w[1]   // 1ustake

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err)

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	sendAmount := 1
	uusdcFee := int64(5)
	_, err = val.ExecTx(ctx, alice.KeyName(), "bank", "send", alice.KeyName(), bob.FormattedAddress(), fmt.Sprintf("%dustake", sendAmount), "--fees", fmt.Sprintf("%duusdc", uusdcFee))
	require.ErrorContains(t, err, "the chain is paused")

	uStakebal, err := noble.GetBalance(ctx, alice.FormattedAddress(), "ustake")
	require.NoError(t, err)
	require.Equal(t, originalAmount, uStakebal)

	uusdcBal, err := noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(mintAmount), uusdcBal)

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Send non TF token but pay fee in TF token while the sender is blacklisted
	// EXPECTED: Request fails

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, alice)

	_, err = val.ExecTx(ctx, alice.KeyName(), "bank", "send", alice.KeyName(), bob.FormattedAddress(), fmt.Sprintf("%dustake", sendAmount), "--fees", fmt.Sprintf("%duusdc", uusdcFee))
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens: unauthorized", alice.FormattedAddress()))

	bal, err := noble.GetBalance(ctx, alice.FormattedAddress(), "ustake")
	require.NoError(t, err)
	require.Equal(t, originalAmount, bal)

	uusdcBal, err = noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(mintAmount), uusdcBal)

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, alice)

	// ACTION: Successfully send non TF token but pay fee in TF token
	// EXPECTED: Success; Fee withdrawn from users balance

	_, err = val.ExecTx(ctx, alice.KeyName(), "bank", "send", alice.KeyName(), bob.FormattedAddress(), fmt.Sprintf("%dustake", sendAmount), "--fees", fmt.Sprintf("%duusdc", uusdcFee))
	require.NoError(t, err)

	bobBalStake, err := noble.GetBalance(ctx, bob.FormattedAddress(), "ustake")
	require.NoError(t, err)
	require.Equal(t, originalAmount.Add(math.NewInt(int64(sendAmount))), bobBalStake)

	aliceBalUusdc, err := noble.GetBalance(ctx, alice.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.EqualValues(t, mintAmount-uusdcFee, aliceBalUusdc.Int64())
}

func TestFiatTFAuthzGrant(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// ACTION: Grant an authz SEND using a TF token while TF is paused
	// EXPECTED: Request fails

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter1 := w[0]
	grantee1 := w[1]

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err := val.AuthzGrant(ctx, granter1, grantee1.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.ErrorContains(t, err, "can not perform token authorizations: the chain is paused")

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Grant an authz SEND using a TF token to a grantee who is blacklisted
	// EXPECTED: Success;

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter2 := w[0]
	grantee2 := w[1]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, grantee2)

	res, err := val.AuthzGrant(ctx, granter2, grantee2.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	// ACTION: Grant an authz SEND using a TF token from a granter who is blacklisted
	// EXPECTED: Success;

	w = interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	granter3 := w[0]
	grantee3 := w[1]

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, granter3)

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

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	// setup
	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble, noble)
	granter := w[0]
	grantee := w[1]
	receiver := w[2]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", granter.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	res, err := val.AuthzGrant(ctx, granter, grantee.FormattedAddress(), "send", "--spend-limit=100uusdc")
	require.NoError(t, err)
	require.Zero(t, res.Code)

	sendAmount := int64(5)
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
	// EXPECTED: Request fails; Even though grantee is acting on behalf of the granter,
	// the granter still cannot execute `send` due to being blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, grantee)

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not authorize tokens: unauthorized", receiver.FormattedAddress()))

	bal, err := noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, bal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, grantee)

	// ACTION: Execute an authz SEND using a TF token from a granter who is blacklisted
	// EXPECTED: Request fails; Granter is blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, granter)

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens: unauthorized", granter.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, bal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, granter)

	// ACTION: Execute an authz SEND using a TF token to a receiver who is blacklisted
	// EXPECTED: Request fails; Granter is blacklisted
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, receiver)

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens: unauthorized", receiver.FormattedAddress()))

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, bal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, receiver)

	// ACTION: Execute an authz SEND using a TF token while the TF is paused
	// EXPECTED: Request fails; chain is paused
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.ErrorContains(t, err, "the chain is paused")

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.True(t, bal.IsZero())

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Happy path: Execute an authz SEND using a TF token
	// EXPECTED: Success; authz send is successful.
	// Status:
	// 	Granter1 has authorized Grantee1 to send 100usdc from their wallet

	_, err = val.AuthzExec(ctx, grantee, nestedCmd)
	require.NoError(t, err)

	bal, err = noble.GetBalance(ctx, receiver.FormattedAddress(), "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(sendAmount), bal)
}

func TestFiatTFBankSend(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, noble)
	alice := w[0]
	bob := w[1]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", alice.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// ACTION: Send TF token while TF is paused
	// EXPECTED: Request fails; token not sent

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

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

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: Send TF token while FROM address is blacklisted
	// EXPECTED: Request fails; token not sent

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, alice)

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens", alice.FormattedAddress()))

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, alice)

	// ACTION: Send TF token while TO address is blacklisted
	// EXPECTED: Request fails; token not sent

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, bob)

	err = noble.SendFunds(ctx, alice.KeyName(), amountToSend)
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens", bob.FormattedAddress()))

	bobBal, err = noble.GetBalance(ctx, bob.FormattedAddress(), "uusdc")
	require.NoError(t, err, "error getting balance")
	require.True(t, bobBal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, bob)

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

	nw, gaia, _, r, ibcPathName, _, eRep, _, _ := e2e.NobleSpinUpIBC(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.OneInt(), noble, gaia)
	nobleWallet := w[0]
	gaiaWallet := w[1]

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nobleWallet.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// noble -> gaia channel info
	nobleToGaiaChannelInfo, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)
	nobleToGaiaChannelID := nobleToGaiaChannelInfo[0].ChannelID
	// gaia -> noble channel info
	gaiaToNobleChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaToNobleChannelID := gaiaToNobleChannelInfo[0].ChannelID

	// uusdc IBC denom on gaia
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaToNobleChannelID, e2e.DenomMetadataUsdc.Base))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	amountToSend := math.NewInt(5)
	transfer := ibc.WalletAmount{
		Address: gaiaWallet.FormattedAddress(),
		Denom:   e2e.DenomMetadataUsdc.Base,
		Amount:  amountToSend,
	}

	// ACTION: IBC send out a TF token while TF is paused
	// EXPECTED: Request fails;

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ibc transfer noble -> gaia
	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, "the chain is paused")

	// relay MsgRecvPacket & MsgAcknowledgement
	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err := gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: IBC send TF token from a blacklisted user
	// EXPECTED: Request fails;

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nobleWallet)

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not send tokens", nobleWallet.FormattedAddress()))

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nobleWallet)

	// ACTION: IBC send out a TF token to a blacklisted user
	// EXPECTED: Request fails;

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, gaiaWallet)

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, fmt.Sprintf("an address (%s) is blacklisted and can not receive tokens", gaiaWallet.FormattedAddress()))

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.True(t, gaiaWalletBal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, gaiaWallet)

	// ACTION: IBC send out a TF token to malformed address
	// EXPECTED: Request fails;

	transfer.Address = "malformed-address1234"

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.ErrorContains(t, err, "error decoding address")

	// ACTION: Successfully IBC send out a TF token to an address on another chain
	// EXPECTED: Success;

	transfer.Address = gaiaWallet.FormattedAddress()

	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	gaiaWalletBal, err = gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, gaiaWalletBal)
}

func TestFiatTFIBCIn(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	nw, gaia, _, r, ibcPathName, _, eRep, _, _ := e2e.NobleSpinUpIBC(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	val := noble.Validators[0]

	w := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.NewInt(1_000_000), noble, gaia)
	nobleWallet := w[0] // 1_000_000ustake
	gaiaWallet := w[1]  // 1_000_000uatom

	mintAmount := int64(100)
	_, err := val.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(), "fiat-tokenfactory", "mint", nobleWallet.FormattedAddress(), fmt.Sprintf("%duusdc", mintAmount))
	require.NoError(t, err, "error minting")

	// noble -> gaia channel info
	nobleToGaiaChannelInfo, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)
	nobleToGaiaChannelID := nobleToGaiaChannelInfo[0].ChannelID
	// gaia -> noble channel info
	gaiaToNobleChannelInfo, err := r.GetChannels(ctx, eRep, gaia.Config().ChainID)
	require.NoError(t, err)
	gaiaToNobleChannelID := gaiaToNobleChannelInfo[0].ChannelID

	amountToSend := math.NewInt(mintAmount)
	transfer := ibc.WalletAmount{
		Address: gaiaWallet.FormattedAddress(),
		Denom:   e2e.DenomMetadataUsdc.Base,
		Amount:  amountToSend,
	}

	// ibc transfer noble -> gaia
	_, err = noble.SendIBCTransfer(ctx, nobleToGaiaChannelID, nobleWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// relay MsgRecvPacket & MsgAcknowledgement
	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, nobleToGaiaChannelID))

	// uusdc IBC denom on gaia
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", gaiaToNobleChannelID, e2e.DenomMetadataUsdc.Base))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	gaiaWalletBal, err := gaia.GetBalance(ctx, gaiaWallet.FormattedAddress(), dstIbcDenom)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, gaiaWalletBal)

	// ACTION: IBC send in a TF token while TF is paused
	// EXPECTED: Request fails;

	e2e.PauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	amountToSend = math.OneInt()
	transfer = ibc.WalletAmount{
		Address: nobleWallet.FormattedAddress(),
		Denom:   dstIbcDenom,
		Amount:  amountToSend,
	}

	height, err := gaia.Height(ctx)
	require.NoError(t, err)

	tx, err := gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err := gaia.Height(ctx)
	require.NoError(t, err)

	ack, err := testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.Contains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err := noble.GetBalance(ctx, nobleWallet.FormattedAddress(), e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)
	require.True(t, nobleWalletBal.IsZero())

	e2e.UnpauseFiatTF(t, ctx, val, nw.FiatTfRoles.Pauser)

	// ACTION: IBC send in a TF token FROM an address that is blacklisted
	// EXPECTED: Request fails;

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, gaiaWallet)

	height, err = gaia.Height(ctx)
	require.NoError(t, err)

	tx, err = gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err = gaia.Height(ctx)
	require.NoError(t, err)

	ack, err = testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.Contains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err = noble.GetBalance(ctx, nobleWallet.FormattedAddress(), e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)
	require.True(t, nobleWalletBal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, gaiaWallet)

	// ACTION: IBC send in a TF token TO an address that is blacklisted
	// EXPECTED: Request fails;

	e2e.BlacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nobleWallet)

	height, err = gaia.Height(ctx)
	require.NoError(t, err)

	tx, err = gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err = gaia.Height(ctx)
	require.NoError(t, err)

	ack, err = testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.Contains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err = noble.GetBalance(ctx, nobleWallet.FormattedAddress(), e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)
	require.True(t, nobleWalletBal.IsZero())

	e2e.UnblacklistAccount(t, ctx, val, nw.FiatTfRoles.Blacklister, nobleWallet)

	// ACTION: Successfully IBC send in a TF token to an address on noble
	// EXPECTED: Success;

	height, err = gaia.Height(ctx)
	require.NoError(t, err)

	tx, err = gaia.SendIBCTransfer(ctx, gaiaToNobleChannelID, gaiaWallet.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err, "error broadcasting IBC send")

	require.NoError(t, r.Flush(ctx, eRep, ibcPathName, gaiaToNobleChannelID))

	heightAfterFlush, err = gaia.Height(ctx)
	require.NoError(t, err)

	ack, err = testutil.PollForAck(ctx, gaia, height, heightAfterFlush+5, tx.Packet)
	require.NoError(t, err, "error polling for ack")
	require.NotContains(t, string(ack.Acknowledgement), "error handling packet")

	nobleWalletBal, err = noble.GetBalance(ctx, nobleWallet.FormattedAddress(), e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, nobleWalletBal)
}
