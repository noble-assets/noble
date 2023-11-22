package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
	"github.com/spf13/cobra"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
>>>>>>> a4ad980 (chore: rename module path (#283))
)

var _ = strconv.Itoa(0)

func CmdUnblacklist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unblacklist [address]",
		Short: "Broadcast message unblacklist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddress := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnblacklist(
				clientCtx.GetFromAddress().String(),
				argAddress,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
