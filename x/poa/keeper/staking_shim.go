package keeper

import (
	"fmt"
	"time"

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
	// TODO
	panic("not yet implemented")
}

func (k *Keeper) Jail(ctx sdk.Context, addr sdk.ConsAddress) {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		panic(fmt.Errorf("validator not found: %s", addr.String()))
	}

	val.JailCount++
	val.JailedTimestamp = time.Now()

	k.SaveValidator(ctx, val)
}

func (k *Keeper) Unjail(ctx sdk.Context, addr sdk.ConsAddress) {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		panic(fmt.Errorf("validator not found: %s", addr.String()))
	}

	val.JailedTimestamp = time.Time{}

	k.SaveValidator(ctx, val)
}

// Delegation allows for getting a particular delegation for a given validator
// and delegator outside the scope of the staking module.
func (k *Keeper) Delegation(ctx sdk.Context, delegator sdk.AccAddress, val sdk.ValAddress) stakingtypes.DelegationI {
	return stakingtypes.NewDelegation(delegator, val, sdk.ZeroDec())
}

// return if the validator is jailed
func (k *Keeper) IsValidatorJailed(ctx sdk.Context, addr sdk.ConsAddress) bool {
	val, found := k.GetValidatorByConsKey(ctx, addr)
	if !found {
		return false
	}

	return val.IsJailed()
}
