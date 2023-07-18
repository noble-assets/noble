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

func (k Keeper) Mint(c context.Context, req *types.QueryGetMintRequest) (*types.QueryGetMintResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMint(ctx, req.SourceDomain, req.SourceDomainSender, req.Nonce)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMintResponse{Mint: val}, nil
}

func (k Keeper) Mints(c context.Context, req *types.QueryAllMintsRequest) (*types.QueryAllMintsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var MintList []types.Mint
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	MintsStore := prefix.NewStore(store, types.MintPrefix(types.MintKeyPrefix))

	pageRes, err := query.Paginate(MintsStore, req.Pagination, func(key []byte, value []byte) error {
		var mint types.Mint
		if err := k.cdc.Unmarshal(value, &mint); err != nil {
			return err
		}

		MintList = append(MintList, mint)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMintsResponse{Mints: MintList, Pagination: pageRes}, nil
}
