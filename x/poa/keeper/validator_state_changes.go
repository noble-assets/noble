package keeper

import (
	"encoding/hex"
	"fmt"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BlockValidatorUpdates calculates the ValidatorUpdates for the current block
// Called in each EndBlock
func (k Keeper) BlockValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	// Calculate validator set changes.

	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	if err != nil {
		panic(err)
	}

	k.Logger(ctx).Info("Returning end block updates", "count", len(updates), "updates", updates)

	return updates
}

// ApplyAndReturnValidatorSetUpdates at the end of every block we update and return the validator set
func (k Keeper) ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
	validators := k.GetAllValidators(ctx)
	activeVals := len(k.GetAllValidatorsInSet(ctx))
	acceptedVals := len(k.GetAllAcceptedValidators(ctx))
	maxVals := k.GetParams(ctx).MaxValidators

	var updates []abci.ValidatorUpdate

	if acceptedVals == 0 {
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
					if existingVal.PubKey.Compare(consKey) == 0 {
						continue ValSetLoop
					}
				}
				// not yet in updates
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
			pubKey, err := cryptocodec.FromTmProtoPublicKey(update.PubKey)
			if err != nil {
				panic(err)
			}
			consAddr := sdk.ConsAddress(pubKey.Address().Bytes())
			val, found := k.GetValidatorByConsKey(ctx, consAddr)
			if !found {
				panic(fmt.Errorf("validator not found by consensus addr: %s", hex.EncodeToString(consAddr)))
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
		if validator.InSet && ((!validator.IsAccepted && acceptedVals > 0) || validator.IsJailed()) {
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

// CalculateValidatorVouch happens at the start of every block to ensure no malacious actors
func (k Keeper) CalculateValidatorVouches(ctx sdk.Context) error {
	qourum := k.GetParams(ctx).Quorum
	acceptedValidators := k.GetAllAcceptedValidators(ctx)
	validators := k.GetAllValidators(ctx)

	// Query method
	for _, validator := range validators {
		vouches := k.GetAllVouchesForValidator(ctx, validator.Address)

		// check the number of vouches are greater that the qourum needed
		if canValidatorJoinConsensus(len(vouches), len(acceptedValidators), qourum) {
			validator.IsAccepted = true
			k.SaveValidator(ctx, validator)
		} else {
			// if the validator does not have enough vouches but is still accepted
			if validator.IsAccepted {
				validator.IsAccepted = false
				k.SaveValidator(ctx, validator)
				if err := k.DeleteAllVouchesByValidator(ctx, validator.Address); err != nil {
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
func canValidatorJoinConsensus(numberOfVouches int, numberOfValidators int, qourum uint32) bool {
	return numberOfValidators > 0 && (float32(numberOfVouches) >= (float32(numberOfValidators))*(float32(qourum)/100))
}
