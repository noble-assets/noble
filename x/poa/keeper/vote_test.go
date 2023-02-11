package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeeperVoteFunctions(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	// Create test vote
	pubKeys := CreateTestPubKeys(2)
	voter, candidate := sdk.ValAddress(pubKeys[0].Address().Bytes()), sdk.ValAddress(pubKeys[1].Address().Bytes())

	vote := &types.Vote{
		VoterAddress:     voter,
		CandidateAddress: candidate,
		InFavor:          true,
	}

	// TODO: split into multiple test cases

	// Set a vote in the store
	keeper.SetVote(ctx, vote)

	// Check the store to see if the vote was saved
	retVal, found := keeper.GetVote(ctx, append(candidate, voter...))
	assert.Equal(t, vote, retVal, "Should return the correct vote from the store")
	require.True(t, found)

	// Get all votes from the store
	allVals := keeper.GetAllVotes(ctx)
	require.Equal(t, 1, len(allVals))

	// Get all votes for a validator from the store
	votesForVal := keeper.GetAllVotesForValidator(ctx, candidate)
	require.Equal(t, 1, len(votesForVal))
}
