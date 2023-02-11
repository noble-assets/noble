package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

func (k Keeper) SetValidator(ctx sdk.Context, validator *types.Validator) {
	k.Set(ctx, validator.Address, types.ValidatorsKey, validator)
}

// NOTE: could this be a state

// SetValidatorIsInSet after a validator has been accepted and not added to the set we need this function
func (k Keeper) SetValidatorIsInSet(ctx sdk.Context, validator *types.Validator, isInSet bool) {
	validator.InSet = isInSet
	k.SetValidator(ctx, validator)
}

// SetValidatorIsAccepted when the validator is accepeted into the consensus accepted is set to true
func (k Keeper) SetValidatorIsAccepted(ctx sdk.Context, validator *types.Validator, isAccepted bool) {
	validator.IsAccepted = isAccepted
	k.SetValidator(ctx, validator)
}

// SetValidatorIsAccepted when the validator is accepeted into the consensus accepted is set to true and is in the validator set
func (k Keeper) SetValidatorIsAcceptedAndInSet(ctx sdk.Context, validator *types.Validator, isAccepted bool, isInSet bool) {
	validator.IsAccepted = isAccepted
	validator.InSet = isInSet
	k.SetValidator(ctx, validator)
}

func (k Keeper) GetValidator(ctx sdk.Context, key []byte) (*types.Validator, bool) {
	val, found := k.Get(ctx, key, types.ValidatorsKey, k.UnmarshalValidator)
	return val.(*types.Validator), found
}

func (k Keeper) UnmarshalValidator(value []byte) (proto.Message, bool) {
	validator := types.Validator{}
	err := k.cdc.UnmarshalInterface(value, &validator)
	if err != nil {
		return &types.Validator{}, false
	}
	return &validator, true
}

// ValidatorSelectorFn allows validators to be selected by certain conditions
type ValidatorSelectorFn func(validator *types.Validator) bool

func (k Keeper) GetAllValidatorsWithCondition(ctx sdk.Context, validatorSelector ValidatorSelectorFn) (validators []*types.Validator) {
	val := k.GetAll(ctx, types.ValidatorsKey, k.UnmarshalValidator)

	// handle the case that there is only one validator in the set
	if len(val) == 1 {
		return append(validators, val[0].(*types.Validator))
	}

	for _, value := range val {
		validator := value.(*types.Validator)
		if validatorSelector(validator) {
			validators = append(validators, validator)
		}
	}

	return validators
}

func (k Keeper) GetAllValidators(ctx sdk.Context) (validators []*types.Validator) {
	var selectAllValidators ValidatorSelectorFn = func(validator *types.Validator) bool {
		return true
	}
	return k.GetAllValidatorsWithCondition(ctx, selectAllValidators)
}

func (k Keeper) GetAllAcceptedValidators(ctx sdk.Context) (validators []*types.Validator) {
	var selectAcceptedValidators ValidatorSelectorFn = func(validator *types.Validator) bool {
		return validator.IsAccepted
	}
	return k.GetAllValidatorsWithCondition(ctx, selectAcceptedValidators)
}

func (k Keeper) GetAllValidatorsAcceptedAndInSet(ctx sdk.Context) (validators []*types.Validator) {
	var selectValidatorsAcceptedAndInSet ValidatorSelectorFn = func(validator *types.Validator) bool {
		return validator.InSet || validator.IsAccepted
	}
	return k.GetAllValidatorsWithCondition(ctx, selectValidatorsAcceptedAndInSet)
}
