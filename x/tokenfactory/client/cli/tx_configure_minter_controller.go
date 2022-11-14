package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

var _ = strconv.Itoa(0)

func CmdConfigureMinterController() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure-minter-controller [controller] [minter]",
		Short: "Broadcast message configure-minter-controller",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argController := args[0]
			argMinter := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgConfigureMinterController(
				clientCtx.GetFromAddress().String(),
				argController,
				argMinter,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
