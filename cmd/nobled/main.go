package main

import (
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"noble/app"
	"noble/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd(
		"noble",
		"cosmos",
		app.DefaultNodeHome,
		"noble-1",
		app.ModuleBasics,
		app.New,
		// this line is used by starport scaffolding # root/arguments
	)
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
