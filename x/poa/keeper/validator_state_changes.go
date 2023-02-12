package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BlockValidatorUpdates calculates the ValidatorUpdates for the current block
// Called in each EndBlock
func (k Keeper) BlockValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	// Calculate validator set changes.
	validatorUpdates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}

	return validatorUpdates
}

// ApplyAndReturnValidatorSetUpdates at the end of every block we update and return the validator set
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	validators := k.GetAllValidators(ctx)
	activeVals := len(k.GetAllValidatorsInSet(ctx))
	maxVals := k.GetParams(ctx).MaxValidators

	var updates []abci.ValidatorUpdate

	if activeVals == 0 {
		// Recovery fallback, active set will be all validators in recent history, regardless of accepted, in set, or jailed.
		history := k.GetAllHistoricalInfo(ctx)
		for _, entry := range history {
		ValSetLoop:
			for _, val := range entry.Valset {
				consKey, err := val.TmConsPublicKey()
				if err != nil {
					panic(err)
				}
				for _, existingVal := range updates {
					if existingVal.PubKey == consKey {
						continue ValSetLoop
					}
					updates = append(updates, abci.ValidatorUpdate{
						PubKey: consKey,
						Power:  types.ValidatorActivePower,
					})
				}
				// not yet in update
				updates = append(updates, abci.ValidatorUpdate{
					PubKey: consKey,
					Power:  types.ValidatorActivePower,
				})
			}
		}

		// If no validators in recent history (i.e. the start of the chain), then active set is all validators
		if len(updates) == 0 {
			for _, val := range validators {
				update, err := val.ABCIValidatorUpdate(types.ValidatorActivePower)
				if err != nil {
					panic(err)
				}
				updates = append(updates, update)
			}
		}

		for _, update := range updates {
			consKey := update.PubKey.GetEd25519()
			pubKey := &ed25519.PubKey{Key: consKey}
			val, found := k.GetValidatorByConsKey(ctx, sdk.ConsAddress(pubKey.Address().Bytes()))
			if !found {
				panic(fmt.Errorf("validator not found by consensus key: %s", hex.EncodeToString(consKey)))
			}
			val.InSet = true
			k.SaveValidator(ctx, val)
		}

		return updates, nil
	}

	addedVals := 0

	for _, validator := range validators {
		// if there are less validators then allowed in the set
		if activeVals+addedVals <= int(maxVals) {
			// validator has been accepted but is not yet in the set
			if !validator.InSet && validator.EligibleToJoinSet() {
				validator.InSet = true
				k.SaveValidator(ctx, validator)
				update, err := validator.ABCIValidatorUpdate(types.ValidatorActivePower)
				if err != nil {
					return nil, err
				}
				addedVals++
				updates = append(updates, update)
			}
		}
		// validator has been kicked but not yet removed from the set
		if validator.InSet && (!validator.IsAccepted || validator.IsJailed()) {
			validator.InSet = false
			k.SaveValidator(ctx, validator)
			update, err := validator.ABCIValidatorUpdate(0)
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

	// Query method
	for _, validator := range validators {
		votes := k.GetAllVotesForValidator(ctx, validator.Address)

		// check the number of votes are greater that the qourum needed
		if canValidatorJoinConsensus(len(votes), len(acceptedValidators), qourum) {
			validator.IsAccepted = true
			k.SaveValidator(ctx, validator)
		} else {
			// if the validator does not have enough votes but is still accepted
			if validator.IsAccepted {
				validator.IsAccepted = false
				k.SaveValidator(ctx, validator)
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
