package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func CmdListPublicKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-public-keys",
		Short: "lists all public keys",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllPublicKeysRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.PublicKeys(context.Background(), params)
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

func CmdShowPublicKey() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-public-key [key]",
		Short: "shows a public key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			key := args[0]

			params := &types.QueryGetPublicKeyRequest{
				Key: key,
			}

			res, err := queryClient.PublicKey(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
