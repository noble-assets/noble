package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func CmdShowMintingDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-minting-denom",
		Short: "shows minting-denom",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetMintingDenomRequest{}

			res, err := queryClient.MintingDenom(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
