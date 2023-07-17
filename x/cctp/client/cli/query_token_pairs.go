package cli

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func CmdListTokenPairs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-token-pairs",
		Short: "lists all token pairs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllTokenPairsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.TokenPairs(context.Background(), params)
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

func CmdShowTokenPair() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-token-pair [remote-domain] [remote-token]",
		Short: "shows a token pair",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			remoteDomain, err := strconv.ParseUint(args[0], 10, 32)
			remoteToken := args[1]
			if err != nil {
				return err
			}

			params := &types.QueryGetTokenPairRequest{
				RemoteDomain: uint32(remoteDomain),
				RemoteToken:  remoteToken,
			}

			res, err := queryClient.TokenPair(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
