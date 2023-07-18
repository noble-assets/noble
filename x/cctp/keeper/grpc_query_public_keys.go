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

func (k Keeper) PublicKey(c context.Context, req *types.QueryGetPublicKeyRequest) (*types.QueryGetPublicKeyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetPublicKey(ctx, req.Key)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetPublicKeyResponse{PublicKey: val}, nil
}

func (k Keeper) PublicKeys(c context.Context, req *types.QueryAllPublicKeysRequest) (*types.QueryAllPublicKeysResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var publicKeys []types.PublicKeys
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	PublicKeysStore := prefix.NewStore(store, types.KeyPrefix(types.PublicKeyKeyPrefix))

	pageRes, err := query.Paginate(PublicKeysStore, req.Pagination, func(key []byte, value []byte) error {
		var publicKey types.PublicKeys
		if err := k.cdc.Unmarshal(value, &publicKey); err != nil {
			return err
		}

		publicKeys = append(publicKeys, publicKey)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllPublicKeysResponse{PublicKeys: publicKeys, Pagination: pageRes}, nil
}
