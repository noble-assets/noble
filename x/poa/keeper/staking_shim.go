package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
)

var _ ibcclienttypes.StakingKeeper = &Keeper{}
var _ evidencetypes.StakingKeeper = &Keeper{}
var _ slashingtypes.StakingKeeper = &Keeper{}

func (k *Keeper) GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool) {
	// TODO
	panic("not yet implemented")
}

func (k *Keeper) UnbondingTime(ctx sdk.Context) time.Duration {
	// TODO
	panic("not yet implemented")
}

func (k *Keeper) ValidatorByConsAddr(ctx sdk.Context, addr sdk.ConsAddress) stakingtypes.ValidatorI {
	// val, found := k.GetValidator(ctx, addr)
	// if !found {
	// 	return nil
	// }

	// return val

	// TODO
	panic("not yet implemented")
}

// iterate through validators by operator address, execute func for each validator
func (k *Keeper) IterateValidators(sdk.Context,
	func(index int64, validator stakingtypes.ValidatorI) (stop bool)) {

}

func (k *Keeper) Validator(sdk.Context, sdk.ValAddress) stakingtypes.ValidatorI {
	// TODO
	panic("not yet implemented")
}

// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
func (k *Keeper) Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec, stakingtypes.InfractionType) {
	// TODO
	panic("not yet implemented")
}

func (k *Keeper) Jail(sdk.Context, sdk.ConsAddress) {
	// TODO
	panic("not yet implemented")
}

func (k *Keeper) Unjail(sdk.Context, sdk.ConsAddress) {
	// TODO
	panic("not yet implemented")
}

// Delegation allows for getting a particular delegation for a given validator
// and delegator outside the scope of the staking module.
func (k *Keeper) Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) stakingtypes.DelegationI {
	// TODO
	panic("not yet implemented")
}

// MaxValidators returns the maximum amount of bonded validators
func (k *Keeper) MaxValidators(sdk.Context) uint32 {
	// TODO
	panic("not yet implemented")
}

// return if the validator is jailed
func (k *Keeper) IsValidatorJailed(ctx sdk.Context, addr sdk.ConsAddress) bool {
	// TODO
	panic("not yet implemented")
}
