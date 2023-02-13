package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/strangelove-ventures/noble/x/poa/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) QueryVouch(c context.Context, req *types.QueryVouchRequest) (*types.QueryVouchResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	candidateAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
	}

	vouchrAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vouchr address as bech32: %w", err)
	}

	val, found := k.GetVouch(ctx, append(candidateAddr, vouchrAddr...))
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryVouchResponse{
		CandidateAddress: req.CandidateAddress,
		VoucherAddress:   req.VoucherAddress,
		InFavor:          val.InFavor,
	}, nil
}

func (k Keeper) QueryVouches(c context.Context, req *types.QueryVouchesRequest) (*types.QueryVouchesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var vouches []*types.QueryVouchResponse
	ctx := sdk.UnwrapSDKContext(c)

	var candidateBz []byte
	if req.CandidateAddress != "" {
		candidateAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
		}
		candidateBz = candidateAddr
	}

	store := ctx.KVStore(k.storeKey)
	vouchStore := prefix.NewStore(store, append(types.VouchesKey, candidateBz...))

	pageRes, err := query.Paginate(vouchStore, req.Pagination, func(key []byte, value []byte) error {
		var vouch types.QueryVouchResponse
		if err := k.cdc.Unmarshal(value, &vouch); err != nil {
			return err
		}

		vouches = append(vouches, &vouch)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVouchesResponse{Vouches: vouches, Pagination: pageRes}, nil
}
