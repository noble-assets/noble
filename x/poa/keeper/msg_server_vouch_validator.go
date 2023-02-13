package keeper

import (
	"context"
	"fmt"

	"github.com/strangelove-ventures/noble/x/poa/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) VouchValidator(goCtx context.Context, msg *types.MsgVouchValidator) (*types.MsgVouchValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	candidateAddr, err := sdk.AccAddressFromBech32(msg.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
	}

	voucherAddr, err := sdk.AccAddressFromBech32(msg.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vouchr address as bech32: %w", err)
	}

	_, found := k.GetValidator(ctx, candidateAddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorFound, fmt.Sprintf("unrecognized %s validator does not exist: %T", types.ModuleName, msg))
	}

	voucher, found := k.GetValidator(ctx, voucherAddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrEmptyValidatorAddr, fmt.Sprintf("unrecognized %s voting validator does not exist: %T", types.ModuleName, msg))
	}

	// if the validator hasn't been accepted they cannot vouch
	if !voucher.IsAccepted {
		return nil, sdkerrors.Wrap(types.ErrNoAcceptedValidatorFound, fmt.Sprintf("error %s validator is not accepted by consensus: %T", types.ModuleName, msg))
	}

	vouch := &types.Vouch{
		VoucherAddress:   voucherAddr,
		CandidateAddress: candidateAddr,
		InFavor:          msg.InFavor,
	}

	k.SetVouch(ctx, vouch)

	// ctx.EventManager().EmitEvents(sdk.Events{
	// 	sdk.NewEvent(
	// 		types.EventTypeVouch,
	// 		sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.Voucher.String()),
	// 		sdk.NewAttribute(types.AttributeKeyCandidate, msg.Name),
	// 	),
	// })

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgVouchValidatorResponse{}, err
}
