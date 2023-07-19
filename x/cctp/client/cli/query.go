package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group cctp queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdListAttesters())
	cmd.AddCommand(CmdListTokenPairs())
	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdShowAuthority())
	cmd.AddCommand(CmdShowBurningAndMintingPaused())
	cmd.AddCommand(CmdShowMaxMessageBodySize())
	cmd.AddCommand(CmdShowMinterAllowance())
	cmd.AddCommand(CmdShowPerMessageBurnLimit())
	cmd.AddCommand(CmdShowAttester())
	cmd.AddCommand(CmdShowSendingAndReceivingMessagesPaused())
	cmd.AddCommand(CmdShowSignatureThreshold())
	cmd.AddCommand(CmdShowTokenPair())

	return cmd
}
