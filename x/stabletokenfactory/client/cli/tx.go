package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/noble-assets/noble/v4/x/stabletokenfactory/types"
	"github.com/spf13/cobra"
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
	cmd.AddCommand(CmdAcceptOwner())
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

	return cmd
}
