package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) UnpauseBurningAndMinting(goCtx context.Context, msg *types.MsgUnpauseBurningAndMinting) (*types.MsgUnpauseBurningAndMintingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot unpause burning and minting")
	}

	// don't unpause if already unpaused
	paused, found := k.GetBurningAndMintingPaused(ctx)
	if found && !paused.Paused {
		return nil, sdkerrors.Wrapf(types.ErrDuringPause, "Burning and minting is already unpaused")
	}

	newPause := types.BurningAndMintingPaused{
		Paused: false,
	}
	k.SetBurningAndMintingPaused(ctx, newPause)

	event := types.UnpauseBurningAndMinting{}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgUnpauseBurningAndMintingResponse{}, err
}
