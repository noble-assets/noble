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
	"encoding/hex"
	"testing"
	"time"

	"cosmossdk.io/math"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func TestCCTP_DepositForBurnWithCaller(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	nobleValidator := noble.Validators[0]

	// SET UP FIAT TOKEN FACTORY AND MINT

	user := interchaintest.GetAndFundTestUsers(t, ctx, "wallet", math.OneInt(), noble)[0]

	_, err := nobleValidator.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress(), nw.FiatTfRoles.Minter.FormattedAddress(),
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", nw.FiatTfRoles.Minter.FormattedAddress(), "1000000000000"+e2e.DenomMetadataUsdc.Base,
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	_, err = nobleValidator.ExecTx(ctx, nw.FiatTfRoles.Minter.KeyName(),
		"fiat-tokenfactory", "mint", user.FormattedAddress(), "1000000000000"+e2e.DenomMetadataUsdc.Base,
	)
	require.NoError(t, err, "failed to execute mint to user tx")

	_, err = nobleValidator.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress(), cctptypes.ModuleAddress.String(),
	)
	require.NoError(t, err, "failed to configure cctp minter controller")

	_, err = nobleValidator.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", cctptypes.ModuleAddress.String(), "1000000000000"+e2e.DenomMetadataUsdc.Base,
	)
	require.NoError(t, err, "failed to configure cctp minter")

	// ----

	broadcaster := cosmos.NewBroadcaster(t, noble)

	burnToken := make([]byte, 32)
	copy(burnToken[12:], common.FromHex("0x07865c6E87B9F70255377e024ace6630C1Eaa37F"))

	tokenMessenger := make([]byte, 32)
	copy(tokenMessenger[12:], common.FromHex("0xD0C3da58f55358142b8d3e06C1C30c5C6114EFE8"))

	bCtx, bCancel := context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err := cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		nw.CCTPRoles.Owner,
		&cctptypes.MsgAddRemoteTokenMessenger{
			From:     nw.CCTPRoles.Owner.FormattedAddress(),
			DomainId: 0,
			Address:  tokenMessenger,
		},
	)
	require.NoError(t, err, "error adding remote token messenger")
	require.Zero(t, tx.Code, "adding remote token messenger failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		nw.CCTPRoles.TokenController,
		&cctptypes.MsgLinkTokenPair{
			From:         nw.CCTPRoles.TokenController.FormattedAddress(),
			RemoteDomain: 0,
			RemoteToken:  burnToken,
			LocalToken:   e2e.DenomMetadataUsdc.Base,
		},
	)
	require.NoError(t, err, "error linking token pair")
	require.Zero(t, tx.Code, "linking token pair failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	beforeBurnBal, err := noble.GetBalance(ctx, user.FormattedAddress(), e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)

	mintRecipient := make([]byte, 32)
	copy(mintRecipient[12:], common.FromHex("0xfCE4cE85e1F74C01e0ecccd8BbC4606f83D3FC90"))

	destinationCaller := []byte("12345678901234567890123456789012")

	msgDepositForBurnWithCallerNoble := &cctptypes.MsgDepositForBurnWithCaller{
		From:              user.FormattedAddress(),
		Amount:            math.NewInt(1000000),
		BurnToken:         e2e.DenomMetadataUsdc.Base,
		DestinationDomain: 0,
		MintRecipient:     mintRecipient,
		DestinationCaller: destinationCaller,
	}

	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		user,
		msgDepositForBurnWithCallerNoble,
	)
	require.NoError(t, err, "error broadcasting msgDepositForBurnWithCaller")
	require.Zero(t, tx.Code, "msgDepositForBurnWithCaller failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	afterBurnBal, err := noble.GetBalance(ctx, user.FormattedAddress(), e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)

	require.Equal(t, afterBurnBal, beforeBurnBal.Sub(math.NewInt(1000000)))

	for _, rawEvent := range tx.Events {
		switch rawEvent.Type {
		case "circle.cctp.v1.DepositForBurn":
			parsedEvent, err := sdk.ParseTypedEvent(rawEvent)
			require.NoError(t, err)
			depositForBurn, ok := parsedEvent.(*cctptypes.DepositForBurn)
			require.True(t, ok)

			expectedBurnToken := hex.EncodeToString(crypto.Keccak256([]byte(e2e.DenomMetadataUsdc.Base)))

			require.Equal(t, uint64(0), depositForBurn.Nonce)
			require.Equal(t, expectedBurnToken, depositForBurn.BurnToken)
			require.Equal(t, msgDepositForBurnWithCallerNoble.Amount, depositForBurn.Amount)
			require.Equal(t, user.FormattedAddress(), depositForBurn.Depositor)
			require.Equal(t, mintRecipient, depositForBurn.MintRecipient)
			require.Equal(t, uint32(0), depositForBurn.DestinationDomain)
			require.Equal(t, tokenMessenger, depositForBurn.DestinationTokenMessenger)
			require.Equal(t, destinationCaller, depositForBurn.DestinationCaller)

		case "circle.cctp.v1.MessageSent":
			parsedEvent, err := sdk.ParseTypedEvent(rawEvent)
			require.NoError(t, err)
			event, ok := parsedEvent.(*cctptypes.MessageSent)
			require.True(t, ok)

			message, err := new(cctptypes.Message).Parse(event.Message)
			require.NoError(t, err)

			messageSender := make([]byte, 32)
			copy(messageSender[12:], sdk.MustAccAddressFromBech32(cctptypes.ModuleAddress.String()))

			expectedBurnToken := crypto.Keccak256([]byte(msgDepositForBurnWithCallerNoble.BurnToken))

			moduleAddress := make([]byte, 32)
			copy(moduleAddress[12:], sdk.MustAccAddressFromBech32(user.FormattedAddress()))

			require.Equal(t, uint32(0), message.Version)
			require.Equal(t, uint32(4), message.SourceDomain)
			require.Equal(t, uint32(0), message.DestinationDomain)
			require.Equal(t, uint64(0), message.Nonce)
			require.Equal(t, messageSender, message.Sender)
			require.Equal(t, tokenMessenger, message.Recipient)
			require.Equal(t, destinationCaller, message.DestinationCaller)

			body, err := new(cctptypes.BurnMessage).Parse(message.MessageBody)
			require.NoError(t, err)

			require.Equal(t, uint32(0), body.Version)
			require.Equal(t, mintRecipient, body.MintRecipient)
			require.Equal(t, msgDepositForBurnWithCallerNoble.Amount, body.Amount)
			require.Equal(t, expectedBurnToken, body.BurnToken)
			require.Equal(t, moduleAddress, body.MessageSender)
		}
	}
}
