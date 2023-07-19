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

func CmdReplaceDepositForBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace-deposit-for-burn [original-message] [original-attestation] [new-destination-caller] [new-mint-recipient]",
		Short: "Broadcast message replace-deposit-for-burn",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			originalMessage := args[0]
			originalAttestation := args[1]
			newDestinationCaller := args[2]
			newMintRecipient := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgReplaceDepositForBurn(
				[]byte(originalMessage),
				[]byte(originalAttestation),
				[]byte(newDestinationCaller),
				[]byte(newMintRecipient),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
