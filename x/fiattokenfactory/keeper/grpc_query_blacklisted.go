package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/strangelove-ventures/noble/v3/x/fiattokenfactory/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlacklistedAll(c context.Context, req *types.QueryAllBlacklistedRequest) (*types.QueryAllBlacklistedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var blacklisteds []types.Blacklisted
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	blacklistedStore := prefix.NewStore(store, types.KeyPrefix(types.BlacklistedKeyPrefix))

	pageRes, err := query.Paginate(blacklistedStore, req.Pagination, func(key []byte, value []byte) error {
		var blacklisted types.Blacklisted
		if err := k.cdc.Unmarshal(value, &blacklisted); err != nil {
			return err
		}

		blacklisteds = append(blacklisteds, blacklisted)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllBlacklistedResponse{Blacklisted: blacklisteds, Pagination: pageRes}, nil
}

func (k Keeper) Blacklisted(c context.Context, req *types.QueryGetBlacklistedRequest) (*types.QueryGetBlacklistedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	_, addressBz, err := bech32.DecodeAndConvert(req.Address)
	if err != nil {
		return nil, err
	}

	val, found := k.GetBlacklisted(ctx, addressBz)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetBlacklistedResponse{Blacklisted: val}, nil
}
