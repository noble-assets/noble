package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

func (k Keeper) SaveValidator(ctx sdk.Context, validator *types.Validator) {
	k.Set(ctx, validator.Address, types.ValidatorsKey, validator)

	pubKey, err := validator.ConsPubKey()
	if err != nil {
		panic(err)
	}
	consAddr := sdk.ConsAddress(pubKey.Address().Bytes())
	k.Set(ctx, consAddr, types.ValidatorsByConsKey, validator)
}

func (k Keeper) GetValidator(ctx sdk.Context, addr sdk.AccAddress) (*types.Validator, bool) {
	val, found := k.Get(ctx, addr, types.ValidatorsKey, k.UnmarshalValidator)
	return val.(*types.Validator), found
}

func (k Keeper) GetValidatorByConsKey(ctx sdk.Context, addr sdk.ConsAddress) (*types.Validator, bool) {
	val, found := k.Get(ctx, addr, types.ValidatorsByConsKey, k.UnmarshalValidator)
	return val.(*types.Validator), found
}

func (k Keeper) UnmarshalValidator(value []byte) (proto.Message, bool) {
	validator := &types.Validator{}
	return validator, validator.Unmarshal(value) == nil
}

// ValidatorSelectorFn allows validators to be selected by certain conditions
type ValidatorSelectorFn func(validator *types.Validator) bool

func (k Keeper) GetAllValidatorsWithCondition(ctx sdk.Context, validatorSelector ValidatorSelectorFn) (validators []*types.Validator) {
	val := k.GetAll(ctx, types.ValidatorsKey, k.UnmarshalValidator)

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

func (k Keeper) GetAllValidatorsStaking(ctx sdk.Context) stakingtypes.Validators {
	vals := k.GetAll(ctx, types.ValidatorsKey, k.UnmarshalValidator)

	var validators stakingtypes.Validators

	for _, value := range vals {
		validator := value.(*types.Validator)
		validators = append(validators, validator.ToStakingValidator())
	}

	return validators
}

func (k Keeper) GetAllAcceptedValidators(ctx sdk.Context) (validators []*types.Validator) {
	var selectAcceptedValidators ValidatorSelectorFn = func(validator *types.Validator) bool {
		return validator.IsAccepted
	}
	return k.GetAllValidatorsWithCondition(ctx, selectAcceptedValidators)
}

func (k Keeper) GetAllValidatorsInSet(ctx sdk.Context) (validators []*types.Validator) {
	var selectAcceptedValidators ValidatorSelectorFn = func(validator *types.Validator) bool {
		return validator.InSet
	}
	return k.GetAllValidatorsWithCondition(ctx, selectAcceptedValidators)
}
