package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

func TestKeeperUpdateValidatorSetFunctions(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	// Create test validators
	pubKeys := CreateTestPubKeys(2)
	valPubKey1, valPubKey2 := pubKeys[0], pubKeys[1]
	valAddr1, valAddr2 := sdk.ValAddress(valPubKey1.Address().Bytes()), sdk.ValAddress(valPubKey2.Address().Bytes())

	pubKeyAny1, err := cdctypes.NewAnyWithValue(valPubKey1)
	require.NoError(t, err)

	pubKeyAny2, err := cdctypes.NewAnyWithValue(valPubKey2)
	require.NoError(t, err)

	validator := &types.Validator{
		Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
		Address:     valAddr1,
		Pubkey:      pubKeyAny1,
	}

	validator2 := &types.Validator{
		Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
		Address:     valAddr2,
		Pubkey:      pubKeyAny2,
	}

	// Set a value in the store
	keeper.SaveValidator(ctx, validator)

	err = keeper.CalculateValidatorVotes(ctx)
	require.NoError(t, err)

	// Validator 1 joins consensus even though it does not have any votes because it's the only validator.
	updates, err := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	keeper.SaveValidator(ctx, validator2)

	vote := &types.Vote{
		VoterAddress:     validator.Address,
		CandidateAddress: validator2.Address,
		InFavor:          true,
	}

	vote2 := &types.Vote{
		VoterAddress:     validator2.Address,
		CandidateAddress: validator2.Address,
		InFavor:          true,
	}

	// Validator 1 votes for validator 2 to join consensus
	keeper.SetVote(ctx, vote)

	// Validator 2 votes for validator 2 to join consensus
	keeper.SetVote(ctx, vote2)

	err = keeper.CalculateValidatorVotes(ctx)
	require.NoError(t, err)

	// Validator 2 joins consensus, but validator 1 is booted because it does not have any votes
	updates, err = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(updates))

	// No updates to the set
	updates, err = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 0, len(updates))
}

func TestKeeperCalculateValidatorVoteFunction(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	// Create test validators
	pubKeys := CreateTestPubKeys(2)
	valPubKey1, valPubKey2 := pubKeys[0], pubKeys[1]
	valAddr1, valAddr2 := sdk.ValAddress(valPubKey1.Address().Bytes()), sdk.ValAddress(valPubKey2.Address().Bytes())

	pubKeyAny1, err := cdctypes.NewAnyWithValue(valPubKey1)
	require.NoError(t, err)

	pubKeyAny2, err := cdctypes.NewAnyWithValue(valPubKey1)
	require.NoError(t, err)

	validator := &types.Validator{
		Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
		Address:     valAddr1,
		Pubkey:      pubKeyAny1,
	}

	validator2 := &types.Validator{
		Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
		Address:     valAddr2,
		Pubkey:      pubKeyAny2,
	}

	// Set a value in the store
	keeper.SaveValidator(ctx, validator)

	_, err = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)

	err = keeper.CalculateValidatorVotes(ctx)
	require.NoError(t, err)

	// Check to see if validator is accepted
	retVal, found := keeper.GetValidator(ctx, validator.Address)
	require.True(t, retVal.IsAccepted)
	require.True(t, found)

	// Set the second validator and assert its not accepted
	keeper.SaveValidator(ctx, validator2)
	retVal, found = keeper.GetValidator(ctx, validator2.Address)
	require.False(t, retVal.IsAccepted)

	vote := &types.Vote{
		VoterAddress:     validator.Address,
		CandidateAddress: validator2.Address,
		InFavor:          true,
	}

	// Validator 1 votes for validator 2 to join the consensus
	keeper.SetVote(ctx, vote)
	err = keeper.CalculateValidatorVotes(ctx)
	require.NoError(t, err)

	// Validator 2's accepted value is set to false
	retVal, found = keeper.GetValidator(ctx, validator2.Address)
	require.True(t, retVal.IsAccepted)
}
