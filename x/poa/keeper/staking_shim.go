package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

var _ ibcclienttypes.StakingKeeper = &Keeper{}
var _ evidencetypes.StakingKeeper = &Keeper{}
var _ slashingtypes.StakingKeeper = &Keeper{}

func (k *Keeper) ValidatorByConsAddr(ctx sdk.Context, addr sdk.ConsAddress) stakingtypes.ValidatorI {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		return nil
	}

	return val
}

// iterate through validators by operator address, execute func for each validator
func (k *Keeper) IterateValidators(ctx sdk.Context,
	fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.ValidatorsKey)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		v, _ := k.UnmarshalValidator(iterator.Value())
		validator := v.(*types.Validator)
		stop := fn(i, validator) // XXX is this safe will the validator unexposed fields be able to get written to?

		if stop {
			break
		}
		i++
	}
}

func (k *Keeper) Validator(ctx sdk.Context, addr sdk.ValAddress) stakingtypes.ValidatorI {
	val, found := k.GetValidator(ctx, addr.Bytes())
	if !found {
		return nil
	}

	return val
}

// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
func (k *Keeper) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight int64, power int64, slashFactor sdk.Dec) {
	logger := k.Logger(ctx)

	if slashFactor.IsNegative() {
		panic(fmt.Errorf("attempted to slash with a negative slash factor: %v", slashFactor))
	}

	// Amount of slashing = slash slashFactor * power at time of infraction
	// amount := sdk.TokensFromConsensusPower(power, k.PowerReduction(ctx))
	// slashAmountDec := amount.ToDec().Mul(slashFactor)
	// slashAmount := slashAmountDec.TruncateInt()

	// ref https://github.com/cosmos/cosmos-sdk/issues/1348

	validator := k.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil {
		// If not found, the validator must have been overslashed and removed - so we don't need to do anything
		// NOTE:  Correctness dependent on invariant that unbonding delegations / redelegations must also have been completely
		//        slashed in this case - which we don't explicitly check, but should be true.
		// Log the slash attempt for future reference (maybe we should tag it too)
		logger.Error(
			"WARNING: ignored attempt to slash a nonexistent validator; we recommend you investigate immediately",
			"validator", consAddr.String(),
		)
		return
	}

	// should not be slashing an unbonded validator
	if validator.IsUnbonded() {
		panic(fmt.Errorf("should not be slashing unbonded validator: %s", validator.GetOperator()))
	}

	// operatorAddress := validator.GetOperator()

	// Track remaining slash amount for the validator
	// This will decrease when we slash unbondings and
	// redelegations, as that stake has since unbonded

	// remainingSlashAmount := slashAmount

	switch {
	case infractionHeight > ctx.BlockHeight():
		// Can't slash infractions in the future
		panic(fmt.Errorf(
			"impossible attempt to slash future infraction at height %d but we are at height %d",
			infractionHeight, ctx.BlockHeight()))

	case infractionHeight == ctx.BlockHeight():
		// Special-case slash at current height for efficiency - we don't need to
		// look through unbonding delegations or redelegations.
		logger.Info(
			"slashing at current height; not scanning unbonding delegations & redelegations",
			"height", infractionHeight,
		)

	case infractionHeight < ctx.BlockHeight():

	}

	// TODO slash tokens from vesting account

	// cannot decrease balance below zero
	// tokensToBurn := sdk.MinInt(remainingSlashAmount, validator.Tokens)
	// tokensToBurn = sdk.MaxInt(tokensToBurn, sdk.ZeroInt()) // defensive.

	// // we need to calculate the *effective* slash fraction for distribution
	// if validator.Tokens.IsPositive() {
	// 	effectiveFraction := tokensToBurn.ToDec().QuoRoundUp(validator.Tokens.ToDec())
	// 	// possible if power has changed
	// 	if effectiveFraction.GT(sdk.OneDec()) {
	// 		effectiveFraction = sdk.OneDec()
	// 	}
	// }

	// // Deduct from validator's bonded tokens and update the validator.
	// // Burn the slashed tokens from the pool account and decrease the total supply.
	// validator = k.RemoveValidatorTokens(ctx, validator, tokensToBurn)

	// switch validator.GetStatus() {
	// case types.Bonded:
	// 	if err := k.burnBondedTokens(ctx, tokensToBurn); err != nil {
	// 		panic(err)
	// 	}
	// case types.Unbonding, types.Unbonded:
	// 	if err := k.burnNotBondedTokens(ctx, tokensToBurn); err != nil {
	// 		panic(err)
	// 	}
	// default:
	// 	panic("invalid validator status")
	// }

	logger.Info(
		"validator slashed by slash factor",
		"validator", validator.GetOperator().String(),
		"slash_factor", slashFactor.String(),
		// "burned", tokensToBurn,
	)
}

func (k *Keeper) Jail(ctx sdk.Context, addr sdk.ConsAddress) {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		panic(fmt.Errorf("validator not found: %s", addr.String()))
	}

	valAddress := sdk.ValAddress(val.Address).String()

	logger := k.Logger(ctx)

	if val.Jailed {
		logger.Error("Attempting to jail a validator that is alread jailed, no-op",
			"validator", valAddress,
		)
		return

	}

	val.JailCount++
	val.Jailed = true

	k.SaveValidator(ctx, val)

	logger.Info("Validator jailed",
		"validator", valAddress,
	)
}

func (k *Keeper) Unjail(ctx sdk.Context, addr sdk.ConsAddress) {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		panic(fmt.Errorf("validator not found: %s", addr.String()))
	}

	valAddress := sdk.ValAddress(val.Address).String()

	logger := k.Logger(ctx)

	if !val.Jailed {
		logger.Error("Attempting to unjail a validator that is not jailed, no-op",
			"validator", valAddress)
		return
	}

	val.Jailed = false

	k.SaveValidator(ctx, val)

	logger.Info("Validator unjailed",
		"validator", valAddress,
	)
}

// Delegation allows for getting a particular delegation for a given validator
// and delegator outside the scope of the module.
func (k *Keeper) Delegation(ctx sdk.Context, delegator sdk.AccAddress, val sdk.ValAddress) stakingtypes.DelegationI {
	//return stakingtypes.NewDelegation(delegator, val, sdk.ZeroDec())
	return nil
}

// return if the validator is jailed
func (k *Keeper) IsValidatorJailed(ctx sdk.Context, addr sdk.ConsAddress) bool {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		return false
	}

	return val.Jailed
}
