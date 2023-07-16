package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func CmdShowPaused() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-paused",
		Short: "shows paused",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetPausedRequest{}

			res, err := queryClient.Paused(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
