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
	"bytes"
	"context"
	_ "embed"
	"path"
	"testing"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

//go:embed data/create_hyperlane_routing_ism.json
var CreateHyperlaneRoutingISM []byte

func TestChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	genesisVersion := "v11.0.1"

	upgrades := []e2e.ChainUpgrade{
		{
			Image:       e2e.LocalImages[0],
			UpgradeName: "citadel",
			PreUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet, _ *e2e.ICATestSuite) {
				val := noble.Validators[0]
				file := "create_hyperlane_routing_ism.json"

				err := val.WriteFile(ctx, CreateHyperlaneRoutingISM, file)
				require.NoError(t, err)

				_, err = val.ExecTx(
					ctx, authority.KeyName(),
					"authority", "execute", path.Join(val.HomeDir(), file),
				)
				require.NoError(t, err)

				stdout, _, err := val.ExecQuery(ctx, "hyperlane", "ism", "isms")
				require.NoError(t, err)
				res := parseHyperlaneISMsResponse(stdout)
				require.Len(t, res.Isms, 1)
			},
			PostUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, _ ibc.Wallet, _ *e2e.ICATestSuite) {
				val := noble.Validators[0]

				stdout, _, err := val.ExecQuery(ctx, "hyperlane", "ism", "isms")
				require.NoError(t, err)
				res := parseHyperlaneISMsResponse(stdout)
				require.Len(t, res.Isms, 2)
			},
		},
	}

	e2e.TestChainUpgrade(t, genesisVersion, upgrades, false)
}

func parseHyperlaneISMsResponse(bz []byte) ismtypes.QueryIsmsResponse {
	var res ismtypes.QueryIsmsResponse
	_ = jsonpb.Unmarshal(bytes.NewReader(bz), &res)

	return res
}
