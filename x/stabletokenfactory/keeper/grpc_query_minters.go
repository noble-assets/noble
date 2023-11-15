package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/strangelove-ventures/noble/v5/x/stabletokenfactory/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MintersAll(c context.Context, req *types.QueryAllMintersRequest) (*types.QueryAllMintersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var minters []types.Minters
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	mintersStore := prefix.NewStore(store, types.KeyPrefix(types.MintersKeyPrefix))

	pageRes, err := query.Paginate(mintersStore, req.Pagination, func(key []byte, value []byte) error {
		var minter types.Minters
		if err := k.cdc.Unmarshal(value, &minter); err != nil {
			return err
		}

		minters = append(minters, minter)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMintersResponse{Minters: minters, Pagination: pageRes}, nil
}

func (k Keeper) Minters(c context.Context, req *types.QueryGetMintersRequest) (*types.QueryGetMintersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMinters(
		ctx,
		req.Address,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMintersResponse{Minters: val}, nil
}
