package cli

import (
	"encoding/binary"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

var _ = strconv.Itoa(0)

func CmdAddTokenMessenger() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-token-messenger [domain-id] [address]",
		Short: "Broadcast message disable-attester",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			domainId := args[0]
			address := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddTokenMessenger(
				clientCtx.GetFromAddress().String(),
				binary.BigEndian.Uint32([]byte(domainId)),
				address,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
