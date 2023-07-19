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

func CmdSendMessageWithCaller() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-message-with-caller [destination-domain] [recipient] [message-body] [destination-caller]",
		Short: "Broadcast message send-message-with-caller",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			destinationDomain, err := strconv.ParseUint(args[0], 10, 32)
			recipient := args[1]
			messageBody := args[2]
			destinationCaller := args[3]

			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSendMessageWithCaller(
				uint32(destinationDomain),
				[]byte(recipient),
				[]byte(messageBody),
				[]byte(destinationCaller),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
