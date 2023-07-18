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

func CmdReplaceMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace-message [original-message] [original-attestation] [new-message-body] [new-destination-caller]",
		Short: "Broadcast message replace-message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			originalMessage := args[0]
			originalAttestation := args[1]
			newMessageBody := args[2]
			newDestinationCaller := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgReplaceMessage(
				[]byte(originalMessage),
				[]byte(originalAttestation),
				[]byte(newMessageBody),
				[]byte(newDestinationCaller),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
