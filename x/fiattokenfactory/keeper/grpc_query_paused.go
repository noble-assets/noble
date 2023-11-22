package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/grpc_query_paused.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/grpc_query_paused.go

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Paused(c context.Context, req *types.QueryGetPausedRequest) (*types.QueryGetPausedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val := k.GetPaused(ctx)

	return &types.QueryGetPausedResponse{Paused: val}, nil
}
