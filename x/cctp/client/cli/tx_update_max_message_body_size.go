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

func CmdUpdateMaxMessageBodySize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-max-message-body-size [size]",
		Short: "Broadcast message update-max-message-body-size",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			size, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateMaxMessageBodySize(
				clientCtx.GetFromAddress().String(),
				uint32(size),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
