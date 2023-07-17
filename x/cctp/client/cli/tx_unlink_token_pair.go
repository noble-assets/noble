package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

var _ = strconv.Itoa(0)

func CmdUnlinkTokenPair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link-token-pair [remote-domain] [remote-token] [local-token]",
		Short: "Broadcast message unlink-token-pair",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			remoteDomain, err := strconv.ParseUint(args[0], 10, 32)
			remoteToken := args[1]
			localToken := args[2]

			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnlinkTokenPair(
				clientCtx.GetFromAddress().String(),
				uint32(remoteDomain),
				remoteToken,
				localToken,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
