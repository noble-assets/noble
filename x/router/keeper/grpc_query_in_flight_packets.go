package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/router/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) InFlightPacket(c context.Context, req *types.QueryGetInFlightPacketRequest) (*types.QueryGetInFlightPacketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetInFlightPacket(ctx, req.ChannelId, req.PortId, req.Sequence)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetInFlightPacketResponse{InFlightPacket: val}, nil
}

func (k Keeper) InFlightPackets(c context.Context, req *types.QueryAllInFlightPacketsRequest) (*types.QueryAllInFlightPacketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var InFlightPackets []types.InFlightPacket
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	InFlightPacketsStore := prefix.NewStore(store, types.InFlightPacketPrefix(types.InFlightPacketKeyPrefix))

	pageRes, err := query.Paginate(InFlightPacketsStore, req.Pagination, func(key []byte, value []byte) error {
		var InFlightPacket types.InFlightPacket
		if err := k.cdc.Unmarshal(value, &InFlightPacket); err != nil {
			return err
		}

		InFlightPackets = append(InFlightPackets, InFlightPacket)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllInFlightPacketsResponse{InFlightPackets: InFlightPackets, Pagination: pageRes}, nil
}
