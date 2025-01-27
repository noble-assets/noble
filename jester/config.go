package jester

import (
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cobra"
)

const (
	defaultJesterGRPC = "localhost:9091"
)

// AppendJesterConfig appends the Jester configuration to app.toml
func AppendJesterConfig(srvCfg *serverconfig.Config) (customAppTemplate string, NobleAppConfig interface{}) {
	type JesterConfig struct {
		GRPCAddress string `mapstructure:"grpc-server"`
	}

	type CustomAppConfig struct {
		serverconfig.Config

		JesterConfig JesterConfig `mapstructure:"jester"`
	}

	defaultJesterConfig := JesterConfig{
		GRPCAddress: defaultJesterGRPC,
	}

	NobleAppConfig = CustomAppConfig{Config: *srvCfg, JesterConfig: defaultJesterConfig}

	customAppTemplate = serverconfig.DefaultConfigTemplate + `
###############################################################################
###                             Jester (sidecar)                            ###
###############################################################################

[jester]

# Jester's gRPC server address. 
# This should not conflict with the CometBFT gRPC server.
grpc-server = "{{ .JesterConfig.GRPCAddress }}"
`
	return customAppTemplate, NobleAppConfig
}

// Flags

const (
	FlagJesterGRPC = "jester.grpc-server"
)

func AddJesterFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagJesterGRPC, defaultJesterGRPC, "Jester's gRPC server address")
}
