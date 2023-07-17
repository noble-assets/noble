package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Attester(c context.Context, req *types.QueryGetAttesterRequest) (*types.QueryGetAttesterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetAttester(ctx, req.Attester)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetAttesterResponse{Attester: val}, nil
}

func (k Keeper) Attesters(c context.Context, req *types.QueryAllAttestersRequest) (*types.QueryAllAttestersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var attesters []types.Attester
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	AttesterStore := prefix.NewStore(store, types.KeyPrefix(types.AttesterKeyPrefix))

	pageRes, err := query.Paginate(AttesterStore, req.Pagination, func(key []byte, value []byte) error {
		var attester types.Attester
		if err := k.cdc.Unmarshal(value, &attester); err != nil {
			return err
		}

		attesters = append(attesters, attester)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAttestersResponse{Attester: attesters, Pagination: pageRes}, nil
}
