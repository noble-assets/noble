package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/strangelove-ventures/noble/app"
	"github.com/strangelove-ventures/noble/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(
		"noble",
		"noble",
		app.DefaultNodeHome,
		"noble-1",
		app.ModuleBasics,
		app.New,
	)

	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
