package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/noble-assets/noble/v5/x/forwarding/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  types.ModuleName,
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(QueryAddress())
	cmd.AddCommand(QueryStats())

	return cmd
}

func QueryAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address [channel] [recipient]",
		Short: "Query forwarding address by channel and recipient",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAddress{Channel: args[0], Recipient: args[1]}

			res, err := queryClient.Address(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func QueryStats() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats [channel]",
		Short: "Query forwarding stats by channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryStatsByChannel{Channel: args[0]}

			res, err := queryClient.StatsByChannel(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
