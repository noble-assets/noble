package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"noble/x/tokenfactory/types"
)

func (k Keeper) MintingDenom(c context.Context, req *types.QueryGetMintingDenomRequest) (*types.QueryGetMintingDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMintingDenom(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMintingDenomResponse{MintingDenom: val}, nil
}
