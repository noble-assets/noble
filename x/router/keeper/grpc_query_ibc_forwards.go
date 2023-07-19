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

func (k Keeper) IBCForward(c context.Context, req *types.QueryGetIBCForwardRequest) (*types.QueryGetIBCForwardResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetIBCForward(ctx, req.SourceDomain, req.SourceDomainSender, req.Nonce)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetIBCForwardResponse{IbcForward: val}, nil
}

func (k Keeper) IBCForwards(c context.Context, req *types.QueryAllIBCForwardsRequest) (*types.QueryAllIBCForwardsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var IBCForwards []types.StoreIBCForwardMetadata
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	IBCForwardsStore := prefix.NewStore(store, types.IBCForwardPrefix(types.IBCForwardKeyPrefix))

	pageRes, err := query.Paginate(IBCForwardsStore, req.Pagination, func(key []byte, value []byte) error {
		var IBCForward types.StoreIBCForwardMetadata
		if err := k.cdc.Unmarshal(value, &IBCForward); err != nil {
			return err
		}

		IBCForwards = append(IBCForwards, IBCForward)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllIBCForwardsResponse{IbcForwards: IBCForwards, Pagination: pageRes}, nil
}
