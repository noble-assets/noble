package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func CmdShowMinterAllowance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-minter-allowance [denom]",
		Short: "shows the minter allowance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			denom := args[0]

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetMinterAllowanceRequest{
				Denom: denom,
			}

			res, err := queryClient.MinterAllowance(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
