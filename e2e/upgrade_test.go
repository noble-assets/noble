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
	_ "embed"
	"path"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"github.com/noble-assets/noble/e2e"
)

//go:embed data/ism.json
var InnerTx []byte

func TestChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	genesisVersion := "v11.0.0-rc.1"

	upgrades := []e2e.ChainUpgrade{
		{
			Image:       e2e.LocalImages[0],
			UpgradeName: "ignition",
			PreUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, _ *e2e.ICATestSuite) {
				validator := noble.Validators[0]
				require.NoError(t, validator.WriteFile(ctx, InnerTx, "ism.json"))
				_, err := validator.ExecTx(
					ctx, authority.KeyName(),
					"authority", "execute", path.Join(validator.HomeDir(), "ism.json"),
				)
				require.NoError(t, err)
			},
		},
	}

	e2e.TestChainUpgrade(t, genesisVersion, upgrades, false)
}
