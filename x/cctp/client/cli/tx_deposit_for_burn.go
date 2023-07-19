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

func CmdDepositForBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit-for-burn [amount] [destination-domain] [mint-recipient] [burn-token]",
		Short: "Broadcast message deposit-for-burn",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			destinationDomain, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return err
			}

			mintRecipient := args[2]
			burnToken := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgDepositForBurn(
				uint32(amount),
				uint32(destinationDomain),
				[]byte(mintRecipient),
				burnToken,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
