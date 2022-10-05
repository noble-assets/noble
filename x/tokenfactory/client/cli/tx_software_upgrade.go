package cli

import (
	"encoding/json"
	"strconv"

	"noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdSoftwareUpgrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "software-upgrade [plan]",
		Short: "Broadcast message software-upgrade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var plan upgradetypes.Plan

			err = json.Unmarshal([]byte(args[0]), &plan)
			if err != nil {
				return err
			}

			msg := types.NewMsgSoftwareUpgrade(
				clientCtx.GetFromAddress().String(),
				plan,
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
