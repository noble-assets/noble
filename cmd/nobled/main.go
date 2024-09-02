package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/noble-assets/noble/v7/app"
	"github.com/noble-assets/noble/v7/cmd"
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
