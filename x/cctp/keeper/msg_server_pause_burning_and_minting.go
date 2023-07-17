package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) PauseBurningAndMinting(goCtx context.Context, msg *types.MsgPauseBurningAndMinting) (*types.MsgPauseBurningAndMintingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot pause burning and minting")
	}

	// don't pause if already paused
	paused, found := k.GetBurningAndMintingPaused(ctx)
	if found && paused.Paused {
		return nil, sdkerrors.Wrapf(types.ErrDuringPause, "Burning and minting is already paused")
	}

	newPause := types.BurningAndMintingPaused{
		Paused: true,
	}
	k.SetBurningAndMintingPaused(ctx, newPause)

	event := types.PauseBurningAndMinting{}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgPauseBurningAndMintingResponse{}, err
}
