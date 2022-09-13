package cli

import (
	"fmt"
	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"noble/x/tokenfactory/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group tokenfactory queries under a subcommand
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
	cmd.AddCommand(CmdShowAdmin())
	// this line is used by starport scaffolding # 1

	return cmd
}
