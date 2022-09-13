package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"noble/x/tokenfactory/types"
)

func (k Keeper) Blacklister(c context.Context, req *types.QueryGetBlacklisterRequest) (*types.QueryGetBlacklisterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetBlacklister(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBlacklisterResponse{Blacklister: val}, nil
}
