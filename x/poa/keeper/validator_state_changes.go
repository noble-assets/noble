package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// ApplyAndReturnValidatorSetUpdates at the end of every block we update and return the validator set
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	validators := k.GetAllValidatorsAcceptedAndInSet(ctx)
	maxVals := k.GetParams(ctx).MaxValidators

	// handle the case if there is only one validator in the set
	if len(validators) == 1 && !validators[0].InSet {
		k.SetValidatorIsAcceptedAndInSet(ctx, validators[0], true, true)
		update, err := validators[0].ABCIValidatorUpdate(k.cdc, 10)
		if err != nil {
			return nil, err
		}
		return []abci.ValidatorUpdate{update}, nil
	}

	var updates []abci.ValidatorUpdate

	for _, validator := range validators {
		// if there are less validators then allowed in the set
		if len(validators) <= int(maxVals) {
			// validator has been accepted but is not yet in the set
			if validator.IsAccepted && !validator.InSet {
				k.SetValidatorIsInSet(ctx, validator, true)
				update, err := validator.ABCIValidatorUpdate(k.cdc, 10)
				if err != nil {
					return nil, err
				}
				updates = append(updates, update)
			}
		}
		// validator has been kicked but not yet removed from the set
		if !validator.IsAccepted && validator.InSet {
			k.SetValidatorIsInSet(ctx, validator, false)
			update, err := validator.ABCIValidatorUpdate(k.cdc, 0)
			if err != nil {
				return nil, err
			}
			updates = append(updates, update)
		}
	}

	return updates, nil
}

// CalculateValidatorVote happens at the start of every block to ensure no malacious actors
func (k Keeper) CalculateValidatorVotes(ctx sdk.Context) error {
	qourum := k.GetParams(ctx).Quorum
	acceptedValidators := k.GetAllAcceptedValidators(ctx)
	validators := k.GetAllValidators(ctx)

	// NOTE: could we add a vote-validator msg to genesis and be able to remove L43:46
	if len(validators) == 1 {
		return nil
	}

	// Query method
	for _, validator := range validators {
		votes := k.GetAllVotesForValidator(ctx, validator.Address)

		// check the number of votes are greater that the qourum needed
		if canValidatorJoinConsensus(len(votes), len(acceptedValidators), qourum) {
			k.SetValidatorIsAccepted(ctx, validator, true)
		} else {
			// if the validator does not have enough votes but is still accepted
			if validator.IsAccepted {
				k.SetValidatorIsAccepted(ctx, validator, false)
				if err := k.DeleteAllVotesByValidator(ctx, validator.Address); err != nil {
					return err
				}
				// TODO: avoid cascading changes to validator set
			}
		}
	}

	// TODO: Jail validators if malicious

	return nil
}

// canValidatorJoinConsensus if this function returns true a validator can join consensus
func canValidatorJoinConsensus(numberOfVotes int, numberOfValidators int, qourum uint32) bool {
	return (float32(numberOfVotes) >= (float32(numberOfValidators))*(float32(qourum)/100))
}
