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

func (k Keeper) MinterAllowance(c context.Context, req *types.QueryGetMinterAllowanceRequest) (*types.QueryGetMinterAllowanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMinterAllowance(ctx, req.Denom)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMinterAllowanceResponse{Allowance: val}, nil
}

func (k Keeper) MinterAllowances(c context.Context, req *types.QueryAllMinterAllowancesRequest) (*types.QueryAllMinterAllowancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var minterAllowances []types.MinterAllowances
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	MinterAllowancesStore := prefix.NewStore(store, types.KeyPrefix(types.MinterAllowanceKeyPrefix))

	pageRes, err := query.Paginate(MinterAllowancesStore, req.Pagination, func(key []byte, value []byte) error {
		var minterAllowance types.MinterAllowances
		if err := k.cdc.Unmarshal(value, &minterAllowance); err != nil {
			return err
		}

		minterAllowances = append(minterAllowances, minterAllowance)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMinterAllowancesResponse{MinterAllowances: minterAllowances, Pagination: pageRes}, nil
}
