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
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/noble-assets/noble/v12/jester"
	"github.com/noble-assets/nova"
)

// AppendConfigs appends Noble's custom configurations to the Cosmos SDK app.toml
func AppendConfigs(config *serverconfig.Config) (customAppTemplate string, customAppConfig interface{}) {
	type CustomAppConfig struct {
		serverconfig.Config

		JesterConfig jester.Config `json:"jester"`
		NovaConfig   nova.Config   `json:"nova"`
	}

	customAppTemplate = serverconfig.DefaultConfigTemplate + jester.ConfigTemplate + nova.ConfigTemplate

	defaultJesterConfig := jester.Config{GRPCAddress: jester.DefaultGRPCAddress}
	defaultNovaConfig := nova.Config{RPCAddress: nova.DefaultRPCAddress}
	customAppConfig = CustomAppConfig{Config: *config, JesterConfig: defaultJesterConfig, NovaConfig: defaultNovaConfig}

	return
}
