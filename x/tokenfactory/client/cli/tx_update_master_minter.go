package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
	"github.com/spf13/cobra"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
=======
>>>>>>> a4ad980 (chore: rename module path (#283))
)

var _ = strconv.Itoa(0)

func CmdUpdateMasterMinter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-master-minter [address]",
		Short: "Broadcast message update-master-minter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddress := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateMasterMinter(
				clientCtx.GetFromAddress().String(),
				argAddress,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
