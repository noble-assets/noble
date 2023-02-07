package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

var _ = strconv.Itoa(0)

func CmdBlacklist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blacklist [address]",
		Short: "Broadcast message blacklist",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddress := args[0]
			_, pubBz, err := bech32.DecodeAndConvert(argAddress)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgBlacklist(
				clientCtx.GetFromAddress().String(),
				pubBz,
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
