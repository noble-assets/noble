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

package jester

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cobra"
)

const (
	DefaultGRPCAddress = "localhost:9091"
	FlagGRPCAddress    = "jester.grpc-address"
)

type Config struct {
	GRPCAddress string `mapstructure:"grpc-address"`
}

const ConfigTemplate = `
###############################################################################
###                             Jester (sidecar)                            ###
###############################################################################

[jester]

# Jester's gRPC server address. 
# This should not conflict with the CometBFT gRPC server.
grpc-address = "{{ .JesterConfig.GRPCAddress }}"
`

// AppendConfig appends the Jester configuration to the Cosmos SDK app.toml
func AppendConfig(config *serverconfig.Config) (customAppTemplate string, customAppConfig interface{}) {
	type CustomAppConfig struct {
		serverconfig.Config

		JesterConfig Config `mapstructure:"jester"`
	}

	customAppTemplate = serverconfig.DefaultConfigTemplate + ConfigTemplate

	defaultJesterConfig := Config{GRPCAddress: DefaultGRPCAddress}
	customAppConfig = CustomAppConfig{Config: *config, JesterConfig: defaultJesterConfig}

	return
}

// AddFlags adds the Jester flags to the default Cosmos SDK start command.
func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagGRPCAddress, DefaultGRPCAddress, "Jester's gRPC Address")
}
