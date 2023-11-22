package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/grpc_query_minting_denom.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/grpc_query_minting_denom.go

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
