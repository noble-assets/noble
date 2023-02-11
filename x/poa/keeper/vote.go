package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

// SetVote sets a vote with key as votee and voter combined in a byte array
func (k Keeper) SetVote(ctx sdk.Context, vote *types.Vote) {
	k.Set(ctx, append(vote.CandidateAddress, vote.VoterAddress...), types.VotesKey, vote)

	k.Set(ctx, append(vote.VoterAddress, vote.CandidateAddress...), types.VotesByValidatorKey, vote)
}

func (k Keeper) GetVote(ctx sdk.Context, key []byte) (types.Vote, bool) {
	vote, found := k.Get(ctx, key, types.VotesKey, k.UnmarshalVote)
	return vote.(types.Vote), found
}

func (k Keeper) DeleteVote(ctx sdk.Context, key []byte) {
	k.Delete(ctx, key, types.VotesKey)
}

func (k Keeper) UnmarshalVote(value []byte) (proto.Message, bool) {
	vote := types.Vote{}
	err := k.cdc.UnmarshalInterface(value, &vote)
	if err != nil {
		return &types.Vote{}, false
	}
	return &vote, true
}

// VoteSelectorFn allows validators to be selected by certain conditions
type VoteSelectorFn func(vote types.Vote) bool

func (k Keeper) GetAllVotesWithCondition(ctx sdk.Context, key []byte, voteSelector VoteSelectorFn) (votes []types.Vote) {
	val := k.GetAll(ctx, key, k.UnmarshalVote)

	for _, value := range val {
		vote := value.(types.Vote)
		if voteSelector(vote) {
			votes = append(votes, value.(types.Vote))
		}
	}

	return votes
}

func (k Keeper) GetAllVotes(ctx sdk.Context) (votes []types.Vote) {
	var selectAllVotes VoteSelectorFn = func(votes types.Vote) bool {
		return true
	}

	return k.GetAllVotesWithCondition(ctx, types.VotesKey, selectAllVotes)
}

func (k Keeper) GetAllVotesForValidator(ctx sdk.Context, validatorAddress []byte) (votes []types.Vote) {
	var selectAllVotesForValidators VoteSelectorFn = func(vote types.Vote) bool {
		return vote.InFavor == true
	}

	return k.GetAllVotesWithCondition(ctx, append(types.VotesKey, validatorAddress...), selectAllVotesForValidators)
}

func (k Keeper) DeleteAllVotesByValidator(ctx sdk.Context, voter []byte) error {

	val := k.GetAll(ctx, append(types.VotesByValidatorKey, voter...), k.UnmarshalVote)

	for _, value := range val {
		vote := value.(types.Vote)

		k.DeleteVote(ctx, append(vote.CandidateAddress, vote.VoterAddress...))
	}

	k.Delete(ctx, voter, types.VotesByValidatorKey)

	return nil
}
