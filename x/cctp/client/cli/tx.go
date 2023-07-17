package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
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

	cmd.AddCommand(CmdAddTokenMessenger())
	cmd.AddCommand(CmdDepositForBurn())
	cmd.AddCommand(CmdDepositForBurnWithCaller())
	cmd.AddCommand(CmdDisableAttester())
	cmd.AddCommand(CmdEnableAttester())
	cmd.AddCommand(CmdLinkTokenPair())
	cmd.AddCommand(CmdPauseBurningAndMinting())
	cmd.AddCommand(CmdPauseSendingAndReceivingMessages())
	cmd.AddCommand(CmdReceiveMessage())
	cmd.AddCommand(CmdRemoveTokenMessenger())
	cmd.AddCommand(CmdReplaceDepositForBurn())
	cmd.AddCommand(CmdReplaceMessage())
	cmd.AddCommand(CmdSendMessage())
	cmd.AddCommand(CmdSendMessageWithCaller())
	cmd.AddCommand(CmdUnlinkTokenPair())
	cmd.AddCommand(CmdUnpauseBurningAndMinting())
	cmd.AddCommand(CmdUnpauseSendingAndReceivingMessages())
	cmd.AddCommand(CmdUpdateAuthority())
	cmd.AddCommand(CmdUpdateMaxMessageBodySize())
	cmd.AddCommand(CmdUpdateMinterAllowance())
	cmd.AddCommand(CmdUpdatePerMessageBurnLimit())
	cmd.AddCommand(CmdUpdateSignatureThreshold())

	return cmd
}
