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
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

func TestChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	genesisVersion := "v7.0.0"

	upgrades := []e2e.ChainUpgrade{
		{
			Image:       e2e.GhcrImage("v8.0.4"),
			UpgradeName: "helium",
			PreUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *e2e.ICATestSuite) {
				icaAddr, err := e2e.RegisterICAAccount(ctx, icaTs)
				require.NoError(t, err, "failed to setup ICA account")
				require.NotEmpty(t, icaAddr, "ICA address should not be empty")

				// After successfully creating and querying the ICA account we need to update the test suite value for later usage.
				icaTs.IcaAddress = icaAddr

				// Assert initial balance of the ICA is correct.
				initBal, err := noble.BankQueryBalance(ctx, icaAddr, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, initBal.Equal(icaTs.InitBal), "invalid balance expected(%s), got(%s)", icaTs.InitBal, initBal)

				// Create and fund a user on Noble for use as the dst address in the bank transfer that we
				// compose below.
				users := interchaintest.GetAndFundTestUsers(t, ctx, "user", icaTs.InitBal, icaTs.Host)
				dstAddress := users[0].FormattedAddress()

				transferAmount := math.NewInt(1_000_000)

				fromAddress := sdk.MustAccAddressFromBech32(icaAddr)
				toAddress := sdk.MustAccAddressFromBech32(dstAddress)
				coin := sdk.NewCoin(icaTs.Host.Config().Denom, transferAmount)
				msgs := []sdk.Msg{banktypes.NewMsgSend(fromAddress, toAddress, sdk.NewCoins(coin))}

				icaTs.Msgs = msgs

				err = e2e.SendICATx(ctx, icaTs)
				require.NoError(t, err, "failed to send ICA tx")

				// Assert that the updated balance of the dst address is correct.
				expectedBal := icaTs.InitBal.Add(transferAmount)
				updatedBal, err := noble.BankQueryBalance(ctx, dstAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, updatedBal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, updatedBal)
			},
			PostUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *e2e.ICATestSuite) {
				msgSend, ok := icaTs.Msgs[0].(*banktypes.MsgSend)
				require.True(t, ok, "expected MsgSend, got %T", icaTs.Msgs[0])

				coin := msgSend.Amount[0]

				// Verify that the previously created ICA no longer works.
				err := e2e.SendICATx(ctx, icaTs)
				require.Error(t, err, "should have failed to send ICA tx after v8.0.0 upgrade")

				// Assert that the balance of the dst address has not updated.
				expectedBal := icaTs.InitBal.Add(coin.Amount)
				bal, err := noble.BankQueryBalance(ctx, msgSend.ToAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, bal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, bal)
			},
		},
		{
			Image:       e2e.LocalImages[0],
			UpgradeName: "argentum",
			PostUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, icaTs *e2e.ICATestSuite) {
				msgSend, ok := icaTs.Msgs[0].(*banktypes.MsgSend)
				require.True(t, ok, "expected MsgSend, got %T", icaTs.Msgs[0])

				coin := msgSend.Amount[0]

				// The ICA tx we sent in the previous PostUpgrade handler will be relayed once the chain restarts with
				// the changes in v8.1 so we need to account for that when asserting the dst address bal is correct.
				expectedBal := icaTs.InitBal.Add(coin.Amount).Add(coin.Amount)
				bal, err := noble.BankQueryBalance(ctx, msgSend.ToAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, bal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, bal)

				// Verify that the previously created ICA works again with new txs as well.
				err = e2e.SendICATx(ctx, icaTs)
				require.NoError(t, err, "failed to send ICA tx")

				// Assert that the balance of the dst address is correct.
				expectedBal = bal.Add(coin.Amount)
				updatedBal, err := noble.BankQueryBalance(ctx, msgSend.ToAddress, noble.Config().Denom)
				require.NoError(t, err, "failed to query bank balance")
				require.True(t, updatedBal.Equal(expectedBal), "invalid balance expected(%s), got(%s)", expectedBal, updatedBal)
			},
		},
	}

	e2e.TestChainUpgrade(t, genesisVersion, upgrades, true)
}
