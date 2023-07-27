package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Pauser(c context.Context, req *types.QueryGetPauserRequest) (*types.QueryGetPauserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetPauser(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPauserResponse{Pauser: val}, nil
}
