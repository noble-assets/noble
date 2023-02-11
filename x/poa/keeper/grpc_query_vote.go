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

func (k Keeper) QueryVote(c context.Context, req *types.QueryVoteRequest) (*types.QueryVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	candidateAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
	}

	voterAddr, err := sdk.AccAddressFromBech32(req.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode voter address as bech32: %w", err)
	}

	val, found := k.GetVote(ctx, append(candidateAddr, voterAddr...))
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryVoteResponse{
		CandidateAddress: req.CandidateAddress,
		VoterAddress:     req.VoterAddress,
		InFavor:          val.InFavor,
	}, nil
}

func (k Keeper) QueryVotes(c context.Context, req *types.QueryVotesRequest) (*types.QueryVotesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var votes []*types.QueryVoteResponse
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
	voteStore := prefix.NewStore(store, append(types.VotesKey, candidateBz...))

	pageRes, err := query.Paginate(voteStore, req.Pagination, func(key []byte, value []byte) error {
		var vote types.QueryVoteResponse
		if err := k.cdc.Unmarshal(value, &vote); err != nil {
			return err
		}

		votes = append(votes, &vote)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVotesResponse{Votes: votes, Pagination: pageRes}, nil
}
