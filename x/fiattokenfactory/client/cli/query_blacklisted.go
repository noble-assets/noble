package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func CmdListBlacklisted() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-blacklisted",
		Short: "list all blacklisted",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllBlacklistedRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.BlacklistedAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowBlacklisted() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-blacklisted [address]",
		Short: "shows a blacklisted",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argAddress := args[0]

			params := &types.QueryGetBlacklistedRequest{
				Address: argAddress,
			}

			res, err := queryClient.Blacklisted(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
