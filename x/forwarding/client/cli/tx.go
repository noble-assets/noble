package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/noble-assets/noble/v4/x/forwarding/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  types.ModuleName,
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(TxRegisterAccount())

	return cmd
}

func TxRegisterAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "register-account [channel] [recipient]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgRegisterAccount{
				Signer:    clientCtx.GetFromAddress().String(),
				Recipient: args[1],
				Channel:   args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
