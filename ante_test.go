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

package noble

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	forwardingtypes "github.com/noble-assets/forwarding/v2/types"
	"github.com/stretchr/testify/require"
)

func TestSigVerificationGasConsumer_ForwardingPubKey(t *testing.T) {
	gasMeter := storetypes.NewGasMeter(1000000)
	initialGas := gasMeter.GasConsumed()

	sig := signing.SignatureV2{
		PubKey: &forwardingtypes.ForwardingPubKey{},
	}

	params := authtypes.DefaultParams()
	err := SigVerificationGasConsumer(gasMeter, sig, params)
	require.NoError(t, err)

	// ForwardingPubKey should not consume any gas
	require.Equal(t, initialGas, gasMeter.GasConsumed())
}

func TestSigVerificationGasConsumer_Secp256k1PubKey(t *testing.T) {
	gasMeter := storetypes.NewGasMeter(1000000)
	initialGas := gasMeter.GasConsumed()

	// Generate a secp256k1 public key for testing
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	sig := signing.SignatureV2{
		PubKey: pubKey,
	}

	params := authtypes.DefaultParams()
	err := SigVerificationGasConsumer(gasMeter, sig, params)
	require.NoError(t, err)

	// Secp256k1 key should consume gas
	gasConsumed := gasMeter.GasConsumed()
	require.Greater(t, gasConsumed, initialGas)
}

func TestSigVerificationGasConsumer_NilPubKey(t *testing.T) {
	gasMeter := storetypes.NewGasMeter(1000000)

	sig := signing.SignatureV2{
		PubKey: nil,
	}

	params := authtypes.DefaultParams()
	err := SigVerificationGasConsumer(gasMeter, sig, params)

	// Nil public key should delegate to default handler which may return error
	// The behavior depends on the default implementation
	_ = err // We just verify it doesn't panic
}
