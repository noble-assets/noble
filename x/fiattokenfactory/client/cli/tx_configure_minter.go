package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/v3/x/fiattokenfactory/types"
)

var _ = strconv.Itoa(0)

func CmdConfigureMinter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure-minter [address] [allowance]",
		Short: "Broadcast message configure-minter",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddress := args[0]
			argAllowance, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgConfigureMinter(
				clientCtx.GetFromAddress().String(),
				argAddress,
				argAllowance,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
