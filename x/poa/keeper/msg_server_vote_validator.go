package keeper

import (
	"context"
	"fmt"

	"github.com/strangelove-ventures/noble/x/poa/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) VoteValidator(goCtx context.Context, msg *types.MsgVoteValidator) (*types.MsgVoteValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	candidateAddr, err := sdk.AccAddressFromBech32(msg.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode candidate address as bech32: %w", err)
	}

	voterAddr, err := sdk.AccAddressFromBech32(msg.CandidateAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode voter address as bech32: %w", err)
	}

	_, found := k.GetValidator(ctx, candidateAddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorFound, fmt.Sprintf("unrecognized %s validator does not exist: %T", types.ModuleName, msg))
	}

	voter, found := k.GetValidator(ctx, voterAddr)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrEmptyValidatorAddr, fmt.Sprintf("unrecognized %s voting validator does not exist: %T", types.ModuleName, msg))
	}

	// if the validator hasn't been accepted they cannot vote
	if !voter.IsAccepted {
		return nil, sdkerrors.Wrap(types.ErrNoAcceptedValidatorFound, fmt.Sprintf("error %s validator is not accepted by consensus: %T", types.ModuleName, msg))
	}

	vote := &types.Vote{
		VoterAddress:     voterAddr,
		CandidateAddress: candidateAddr,
		InFavor:          msg.InFavor,
	}

	k.SetVote(ctx, vote)

	// ctx.EventManager().EmitEvents(sdk.Events{
	// 	sdk.NewEvent(
	// 		types.EventTypeVote,
	// 		sdk.NewAttribute(stakingtypes.AttributeKeyValidator, msg.Voter.String()),
	// 		sdk.NewAttribute(types.AttributeKeyCandidate, msg.Name),
	// 	),
	// })

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgVoteValidatorResponse{}, err
}
