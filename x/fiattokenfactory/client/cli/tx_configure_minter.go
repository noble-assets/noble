package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
	"github.com/spf13/cobra"
<<<<<<< HEAD:x/fiattokenfactory/client/cli/tx_configure_minter.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/client/cli/tx_configure_minter.go
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
