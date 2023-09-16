package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/v4/x/fiattokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MintingDenom(c context.Context, req *types.QueryGetMintingDenomRequest) (*types.QueryGetMintingDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val := k.GetMintingDenom(ctx)

	return &types.QueryGetMintingDenomResponse{MintingDenom: val}, nil
}
