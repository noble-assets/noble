package cli

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"github.com/strangelove-ventures/noble/x/router/types"
)

func CmdListInFlightPackets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-in-flight-packets",
		Short: "lists all in flight packets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllInFlightPacketsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.InFlightPackets(context.Background(), params)
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

func CmdShowInFlightPacket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-in-flight-packet [channel-id] [port-id] [sequence]",
		Short: "shows an in flight packet",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			channelId := args[0]
			portId := args[1]
			sequence, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			params := &types.QueryGetInFlightPacketRequest{
				ChannelId: channelId,
				PortId:    portId,
				Sequence:  sequence,
			}

			res, err := queryClient.InFlightPacket(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
