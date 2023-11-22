package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
	"github.com/spf13/cobra"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/types"
=======
>>>>>>> a4ad980 (chore: rename module path (#283))
)

var _ = strconv.Itoa(0)

func CmdAcceptOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept-owner [address]",
		Short: "Broadcast message accept-owner",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAcceptOwner(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
