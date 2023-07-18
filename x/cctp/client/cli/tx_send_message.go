package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

var _ = strconv.Itoa(0)

func CmdSendMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-message [destination-domain] [recipient] [message-body]",
		Short: "Broadcast message send-message",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			destinationDomain, err := strconv.ParseUint(args[0], 10, 32)
			recipient := args[1]
			messageBody := args[2]

			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSendMessage(
				uint32(destinationDomain),
				[]byte(recipient),
				[]byte(messageBody),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
