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
	"fmt"
	"testing"

	"github.com/noble-assets/noble/e2e"
	"github.com/stretchr/testify/require"
)

func TestRestrictedModules(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	nw, _ := e2e.NobleSpinUp(t, ctx, e2e.LocalImages, false)
	noble := nw.Chain.GetNode()

	restrictedModules := []string{"circuit", "gov", "group"}

	for _, module := range restrictedModules {
		require.False(t, noble.HasCommand(ctx, "query", module), fmt.Sprintf("%s is a restricted module", module))
	}
}
