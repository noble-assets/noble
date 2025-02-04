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
	defaultJesterAddress = "localhost:9091"
)

// AppendJesterConfig appends the Jester configuration to app.toml
func AppendJesterConfig(srvCfg *serverconfig.Config) (customAppTemplate string, NobleAppConfig interface{}) {
	type JesterConfig struct {
		GRPCAddress string `mapstructure:"grpc-address"`
	}

	type CustomAppConfig struct {
		serverconfig.Config

		JesterConfig JesterConfig `mapstructure:"jester"`
	}

	defaultJesterConfig := JesterConfig{
		GRPCAddress: defaultJesterAddress,
	}

	NobleAppConfig = CustomAppConfig{Config: *srvCfg, JesterConfig: defaultJesterConfig}

	customAppTemplate = serverconfig.DefaultConfigTemplate + `
###############################################################################
###                             Jester (sidecar)                            ###
###############################################################################

[jester]

# Jester's gRPC server address. 
# This should not conflict with the CometBFT gRPC server.
grpc-address = "{{ .JesterConfig.GRPCAddress }}"
`
	return customAppTemplate, NobleAppConfig
}

// Flags

const (
	FlagGRPCAddress = "jester.grpc-address"
)

func AddFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagGRPCAddress, defaultJesterAddress, "Jester's gRPC server address")
}
