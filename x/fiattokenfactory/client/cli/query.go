package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/v4/x/fiattokenfactory/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group fiattokenfactory queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListBlacklisted())
	cmd.AddCommand(CmdShowBlacklisted())
	cmd.AddCommand(CmdShowPaused())
	cmd.AddCommand(CmdShowMasterMinter())
	cmd.AddCommand(CmdListMinters())
	cmd.AddCommand(CmdShowMinters())
	cmd.AddCommand(CmdShowPauser())
	cmd.AddCommand(CmdShowBlacklister())
	cmd.AddCommand(CmdShowOwner())
	cmd.AddCommand(CmdListMinterController())
	cmd.AddCommand(CmdShowMinterController())
	cmd.AddCommand(CmdShowMintingDenom())
	// this line is used by starport scaffolding # 1

	return cmd
}
