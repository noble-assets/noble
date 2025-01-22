// Copyright 2024 NASD Inc. All Rights Reserved.
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
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"sort"
	"testing"
	"time"

	"cosmossdk.io/math"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"
)

func TestCCTP_ReceiveMessage(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()
	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, true)
	noble := nw.Chain
	nobleValidator := noble.Validators[0]

	broadcaster := cosmos.NewBroadcaster(t, noble)

	attesters := make([]*ecdsa.PrivateKey, 2)

	// attester - ECDSA public key (Circle will own these keys for mainnet)
	for i := range attesters {
		p, err := crypto.GenerateKey() // private key
		require.NoError(t, err)

		attesters[i] = p

		pubKey := elliptic.Marshal(p.PublicKey, p.PublicKey.X, p.PublicKey.Y) // public key

		attesterPub := hex.EncodeToString(pubKey)

		bCtx, bCancel := context.WithTimeout(ctx, 20*time.Second)
		defer bCancel()

		// Adding an attester to protocol
		tx, err := cosmos.BroadcastTx(
			bCtx,
			broadcaster,
			nw.CCTPRoles.AttesterManager,
			&cctptypes.MsgEnableAttester{
				From:     nw.CCTPRoles.AttesterManager.FormattedAddress(),
				Attester: attesterPub,
			},
		)
		require.NoError(t, err, "error enabling attester")
		require.Zero(t, tx.Code, "cctp enable attester transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)
	}

	burnToken := make([]byte, 32)
	copy(burnToken[12:], common.FromHex("0x07865c6E87B9F70255377e024ace6630C1Eaa37F"))

	bCtx, bCancel := context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err := cosmos.BroadcastTx(
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
	require.Zero(t, tx.Code, "cctp link token pair transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	tokenMessenger := make([]byte, 32)
	copy(tokenMessenger[12:], common.FromHex("0xBd3fa81B58Ba92a82136038B25aDec7066af3155"))

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	tx, err = cosmos.BroadcastTx(
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
	require.Zero(t, tx.Code, "cctp adding remote token messenger transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	_, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()

	cctpModuleAccount := authtypes.NewModuleAddress(cctptypes.ModuleName).String()

	_, err = nobleValidator.ExecTx(ctx, nw.FiatTfRoles.MasterMinter.KeyName(),
		"fiat-tokenfactory", "configure-minter-controller", nw.FiatTfRoles.MinterController.FormattedAddress(), cctpModuleAccount,
	)
	require.NoError(t, err, "failed to execute configure minter controller tx")

	_, err = nobleValidator.ExecTx(ctx, nw.FiatTfRoles.MinterController.KeyName(),
		"fiat-tokenfactory", "configure-minter", cctpModuleAccount, "1000000"+e2e.DenomMetadataUsdc.Base,
	)
	require.NoError(t, err, "failed to execute configure minter tx")

	const receiver = "9B6CA0C13EB603EF207C4657E1E619EF531A4D27" // account

	receiverBz, err := hex.DecodeString(receiver)
	require.NoError(t, err)

	nobleReceiver, err := bech32.ConvertAndEncode(nw.Chain.Config().Bech32Prefix, receiverBz)
	require.NoError(t, err)

	burnRecipientPadded := append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, receiverBz...)

	// someone burned USDC on Ethereum -> Mint on Noble
	depositForBurn := cctptypes.BurnMessage{
		BurnToken:     burnToken,
		MintRecipient: burnRecipientPadded,
		Amount:        math.NewInt(1000000),
		MessageSender: burnRecipientPadded,
	}

	depositForBurnBz, err := depositForBurn.Bytes()
	require.NoError(t, err)

	emptyDestinationCaller := make([]byte, 32)

	wrappedDepositForBurn := cctptypes.Message{
		Version:           0,
		SourceDomain:      0,
		DestinationDomain: 4, // Noble is 4
		Nonce:             0, // dif per message
		Sender:            tokenMessenger,
		Recipient:         cctptypes.PaddedModuleAddress,
		DestinationCaller: emptyDestinationCaller,
		MessageBody:       depositForBurnBz,
	}

	wrappedDepositForBurnBz, err := wrappedDepositForBurn.Bytes()
	require.NoError(t, err)

	digestBurn := crypto.Keccak256(wrappedDepositForBurnBz) // hashed message is the key to the attestation

	attestationBurn := make([]byte, 0, len(attesters)*65) // 65 byte

	// CCTP requires attestations to have signatures sorted by address
	sort.Slice(attesters, func(i, j int) bool {
		return bytes.Compare(
			crypto.PubkeyToAddress(attesters[i].PublicKey).Bytes(),
			crypto.PubkeyToAddress(attesters[j].PublicKey).Bytes(),
		) < 0
	})

	for i := range attesters {
		sig, err := crypto.Sign(digestBurn, attesters[i])
		require.NoError(t, err)

		attestationBurn = append(attestationBurn, sig...)
	}

	t.Logf("Attested to messages: %s", tx.TxHash)

	bCtx, bCancel = context.WithTimeout(ctx, 20*time.Second)
	defer bCancel()
	tx, err = cosmos.BroadcastTx(
		bCtx,
		broadcaster,
		nw.CCTPRoles.Owner,
		&cctptypes.MsgReceiveMessage{ // note: all messages that go to noble go through MsgReceiveMessage
			From:        nw.CCTPRoles.Owner.FormattedAddress(),
			Message:     wrappedDepositForBurnBz,
			Attestation: attestationBurn,
		},
	)
	require.NoError(t, err, "error submitting cctp burn recv tx")
	require.Zerof(t, tx.Code, "cctp burn recv transaction failed: %s - %s - %s", tx.Codespace, tx.RawLog, tx.Data)

	t.Logf("CCTP burn message successfully received: %s", tx.TxHash)

	balance, err := noble.GetBalance(ctx, nobleReceiver, e2e.DenomMetadataUsdc.Base)
	require.NoError(t, err)

	require.Equal(t, math.NewInt(1000000), balance)
}
