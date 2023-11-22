package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
	"github.com/spf13/cobra"
<<<<<<< HEAD:x/fiattokenfactory/client/cli/tx_accept_owner.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/client/cli/tx_accept_owner.go
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
