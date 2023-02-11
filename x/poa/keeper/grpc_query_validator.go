package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/strangelove-ventures/noble/x/poa/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryValidator(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	val, found := k.GetValidator(ctx, addr)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryValidatorResponse{
		ValidatorAddress: req.ValidatorAddress,
		IsAccepted:       val.IsAccepted,
	}, nil
}

func (k Keeper) QueryValidators(c context.Context, req *types.QueryValidatorsRequest) (*types.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var validators []*types.QueryValidatorResponse
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	validatorStore := prefix.NewStore(store, types.ValidatorsKey)

	pageRes, err := query.Paginate(validatorStore, req.Pagination, func(key []byte, value []byte) error {
		var validator types.QueryValidatorResponse
		if err := k.cdc.Unmarshal(value, &validator); err != nil {
			return err
		}

		validators = append(validators, &validator)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryValidatorsResponse{Validators: validators, Pagination: pageRes}, nil
}
