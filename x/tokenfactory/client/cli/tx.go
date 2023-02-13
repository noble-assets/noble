package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdUpdateMasterMinter())
	cmd.AddCommand(CmdUpdatePauser())
	cmd.AddCommand(CmdUpdateBlacklister())
	cmd.AddCommand(CmdUpdateOwner())
	cmd.AddCommand(CmdConfigureMinter())
	cmd.AddCommand(CmdRemoveMinter())
	cmd.AddCommand(CmdMint())
	cmd.AddCommand(CmdBurn())
	cmd.AddCommand(CmdBlacklist())
	cmd.AddCommand(CmdUnblacklist())
	cmd.AddCommand(CmdPause())
	cmd.AddCommand(CmdUnpause())
	cmd.AddCommand(CmdConfigureMinterController())
	cmd.AddCommand(CmdRemoveMinterController())
	// this line is used by starport scaffolding # 1

	return cmd
}
