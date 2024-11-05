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
	"context"
	"testing"

	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
)

func TestConformance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	var nw e2e.NobleWrapper
	nw, ibcSimd, rf, r, ibcPathName, rep, _, client, network := e2e.NobleSpinUpIBC(t, ctx, false)

	conformance.TestChainPair(t, ctx, client, network, nw.Chain, ibcSimd, rf, rep, r, ibcPathName)
}
