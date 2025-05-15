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
	"strconv"
	"testing"

	"cosmossdk.io/math"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/noble-assets/noble/e2e"
)

func TestRateLimit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	const (
		maxPercentSend    = 10
		maxPercentReceive = 10
	)

	faucet := interchaintest.FaucetAccountKeyName

	ctx := context.Background()

	nw, sim, _, r, _, _, eRep, _, _ := e2e.NobleSpinUpIBC(t, ctx, e2e.LocalImages, false)
	noble := nw.Chain
	val := noble.Validators[0]

	simWallet, err := sim.BuildWallet(ctx, "default", "")
	require.NoError(t, err)

	// noble -> ibcSimd channel info
	nobleToSimChannelInfo, err := r.GetChannels(ctx, eRep, noble.Config().ChainID)
	require.NoError(t, err)

	// add rate limit via authority module
	_, err = val.ExecTx(
		ctx, nw.Authority.FormattedAddress(),
		"authority",
		"add-rate-limit",
		noble.Config().Denom,
		nobleToSimChannelInfo[0].ChannelID,
		strconv.Itoa(maxPercentSend),
		strconv.Itoa(maxPercentReceive),
		"24",
	)
	require.NoError(t, err, "failed to execute rate limit tx")

	// get total supply of token to calculate rate limiting thresholds
	totalSupply, err := noble.BankQueryTotalSupplyOf(ctx, noble.Config().Denom)
	require.NoError(t, err)

	// calculate current threshold
	currentSendThreshold := totalSupply.Amount.Mul(math.NewInt(maxPercentSend)).Quo(math.NewInt(100))

	// send up to the the current threshold, this should succeed
	transfer := ibc.WalletAmount{
		Address: simWallet.FormattedAddress(),
		Denom:   noble.Config().Denom,
		Amount:  currentSendThreshold,
	}

	ibcTx, err := noble.SendIBCTransfer(ctx, nobleToSimChannelInfo[0].ChannelID, faucet, transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, ibcTx.Validate(), "failed to validate ibc tx")

	// send over the current threshold, this should fail
	transfer.Amount = math.NewInt(1)
	_, err = noble.SendIBCTransfer(ctx, nobleToSimChannelInfo[0].ChannelID, faucet, transfer, ibc.TransferOptions{})
	require.Error(t, err, "expected rate limit to be hit, but tx was successful")

	// remove rate limit via authority module
	_, err = val.ExecTx(
		ctx, nw.Authority.FormattedAddress(),
		"authority",
		"remove-rate-limit",
		noble.Config().Denom,
		nobleToSimChannelInfo[0].ChannelID,
	)
	require.NoError(t, err, "failed to execute remove rate limit tx")

	// retry sending over the threshold, this should now succeed since rate limit is removed
	ibcTx, err = noble.SendIBCTransfer(ctx, nobleToSimChannelInfo[0].ChannelID, faucet, transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, ibcTx.Validate(), "failed to validate ibc tx")
}
