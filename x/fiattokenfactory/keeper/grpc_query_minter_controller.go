package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/grpc_query_minter_controller.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/grpc_query_minter_controller.go

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MinterControllerAll(c context.Context, req *types.QueryAllMinterControllerRequest) (*types.QueryAllMinterControllerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var minterControllers []types.MinterController
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	minterControllerStore := prefix.NewStore(store, types.KeyPrefix(types.MinterControllerKeyPrefix))

	pageRes, err := query.Paginate(minterControllerStore, req.Pagination, func(key []byte, value []byte) error {
		var minterController types.MinterController
		if err := k.cdc.Unmarshal(value, &minterController); err != nil {
			return err
		}

		minterControllers = append(minterControllers, minterController)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMinterControllerResponse{MinterController: minterControllers, Pagination: pageRes}, nil
}

func (k Keeper) MinterController(c context.Context, req *types.QueryGetMinterControllerRequest) (*types.QueryGetMinterControllerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMinterController(
		ctx,
		req.ControllerAddress,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMinterControllerResponse{MinterController: val}, nil
}
