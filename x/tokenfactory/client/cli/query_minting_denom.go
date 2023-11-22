package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
	"github.com/spf13/cobra"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/types"
=======
>>>>>>> a4ad980 (chore: rename module path (#283))
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
