package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/strangelove-ventures/noble/v4/app"
	"github.com/strangelove-ventures/noble/v4/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(
		app.Name,
		app.AccountAddressPrefix,
		app.DefaultNodeHome,
		app.ChainID,
		app.ModuleBasics,
		app.New,
	)

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
