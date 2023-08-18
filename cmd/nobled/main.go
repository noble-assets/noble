package main

import (
	"errors"
	"io"
	"os"

	cmtDb "github.com/cometbft/cometbft-db"
	cmtCfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/server"
	serverCmd "github.com/cosmos/cosmos-sdk/server/cmd"
	serverCfg "github.com/cosmos/cosmos-sdk/server/config"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/testing/simapp/params"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/app"

	// Auth
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	// Crisis
	"github.com/cosmos/cosmos-sdk/x/crisis"
	// GenUtil
	genUtilCmd "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func init() {
	cfg := sdk.GetConfig()
	cfg.Seal()
}

func main() {
	rootCmd := NewRootCmd()

	if err := serverCmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		var e server.ErrorCode
		switch {
		case errors.As(err, &e):
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}

func NewRootCmd() *cobra.Command {
	tempApp := app.NewNobleApp(log.NewNopLogger(), cmtDb.NewMemDB(), nil, true, sims.NewAppOptionsWithFlagHome(app.DefaultNodeHome))
	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Marshaler:         tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authTypes.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   "simd",
		Short: "simulation app",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			appConfig := serverCfg.DefaultConfig()
			appConfig.MinGasPrices = "0stake"
			return server.InterceptConfigsPreRunHandler(cmd, serverCfg.DefaultConfigTemplate, appConfig, cmtCfg.DefaultConfig())
		},
	}

	server.AddCommands(
		rootCmd,
		app.DefaultNodeHome,
		createApp,
		exportApp,
		// Required for interchaintest.
		crisis.AddModuleInitFlags,
	)

	rootCmd.AddCommand(
		config.Cmd(),
		debug.Cmd(),
		genUtilCmd.GenesisCoreCommand(encodingConfig.TxConfig, app.ModuleBasics, app.DefaultNodeHome),
		genUtilCmd.InitCmd(app.ModuleBasics, app.DefaultNodeHome),
		keys.Commands(app.DefaultNodeHome),
		queryCommand(),
		txCommand(),
	)

	return rootCmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Query subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	app.ModuleBasics.AddQueryCommands(cmd)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transaction subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	app.ModuleBasics.AddTxCommands(cmd)

	return cmd
}

func createApp(
	logger log.Logger, db cmtDb.DB, traceStore io.Writer,
	appOpts serverTypes.AppOptions,
) serverTypes.Application {
	return app.NewNobleApp(
		logger,
		db,
		traceStore,
		true,
		appOpts,
		server.DefaultBaseappOptions(appOpts)...,
	)
}

func exportApp(
	_ log.Logger, _ cmtDb.DB, _ io.Writer, _ int64, _ bool, _ []string,
	_ serverTypes.AppOptions, _ []string,
) (serverTypes.ExportedApp, error) {
	panic("UNIMPLEMENTED")
}
