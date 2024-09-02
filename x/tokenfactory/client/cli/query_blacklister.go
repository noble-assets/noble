package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/noble-assets/noble/v7/x/tokenfactory/types"
	"github.com/spf13/cobra"
)

func CmdShowBlacklister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-blacklister",
		Short: "shows blacklister",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetBlacklisterRequest{}

			res, err := queryClient.Blacklister(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
