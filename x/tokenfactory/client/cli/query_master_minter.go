package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

func CmdShowMasterMinter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-master-minter",
		Short: "shows master-minter",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetMasterMinterRequest{}

			res, err := queryClient.MasterMinter(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
