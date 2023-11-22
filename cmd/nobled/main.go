package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/app"
	"github.com/strangelove-ventures/noble/cmd"
=======
	"github.com/noble-assets/noble/v5/app"
	"github.com/noble-assets/noble/v5/cmd"
>>>>>>> a4ad980 (chore: rename module path (#283))
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
