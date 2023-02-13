package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

// SetVouch sets a vouch with key as vouche and vouchr combined in a byte array
func (k Keeper) SetVouch(ctx sdk.Context, vouch *types.Vouch) {
	k.Set(ctx, append(vouch.CandidateAddress, vouch.VoucherAddress...), types.VouchesKey, vouch)

	k.Set(ctx, append(vouch.VoucherAddress, vouch.CandidateAddress...), types.VouchesByValidatorKey, vouch)
}

func (k Keeper) GetVouch(ctx sdk.Context, key []byte) (*types.Vouch, bool) {
	vouch, found := k.Get(ctx, key, types.VouchesKey, k.UnmarshalVouch)
	if !found || vouch == nil {
		return nil, false
	}
	return vouch.(*types.Vouch), found
}

func (k Keeper) DeleteVouch(ctx sdk.Context, key []byte) {
	k.Delete(ctx, key, types.VouchesKey)
}

func (k Keeper) UnmarshalVouch(value []byte) (proto.Message, bool) {
	if value == nil {
		return nil, false
	}
	vouch := &types.Vouch{}
	return vouch, vouch.Unmarshal(value) == nil
}

// VouchSelectorFn allows validators to be selected by certain conditions
type VouchSelectorFn func(vouch *types.Vouch) bool

func (k Keeper) GetAllVouchesWithCondition(ctx sdk.Context, key []byte, vouchSelector VouchSelectorFn) (vouches []*types.Vouch) {
	val := k.GetAll(ctx, key, k.UnmarshalVouch)

	for _, value := range val {
		vouch := value.(*types.Vouch)
		if vouchSelector(vouch) {
			vouches = append(vouches, value.(*types.Vouch))
		}
	}

	return vouches
}

func (k Keeper) GetAllVouches(ctx sdk.Context) (vouches []*types.Vouch) {
	var selectAllVouches VouchSelectorFn = func(vouches *types.Vouch) bool {
		return true
	}

	return k.GetAllVouchesWithCondition(ctx, types.VouchesKey, selectAllVouches)
}

func (k Keeper) GetAllVouchesForValidator(ctx sdk.Context, validatorAddress []byte) (vouches []*types.Vouch) {
	var selectAllVouchesForValidators VouchSelectorFn = func(vouch *types.Vouch) bool {
		return vouch.InFavor
	}

	return k.GetAllVouchesWithCondition(ctx, append(types.VouchesKey, validatorAddress...), selectAllVouchesForValidators)
}

func (k Keeper) DeleteAllVouchesByValidator(ctx sdk.Context, vouchr []byte) error {

	val := k.GetAll(ctx, append(types.VouchesByValidatorKey, vouchr...), k.UnmarshalVouch)

	for _, value := range val {
		vouch := value.(*types.Vouch)

		k.DeleteVouch(ctx, append(vouch.CandidateAddress, vouch.VoucherAddress...))
	}

	k.Delete(ctx, vouchr, types.VouchesByValidatorKey)

	return nil
}
