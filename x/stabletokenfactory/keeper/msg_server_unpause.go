package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/v5/x/stabletokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Unpause(goCtx context.Context, msg *types.MsgUnpause) (*types.MsgUnpauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pauser, found := k.GetPauser(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "pauser is not set")
	}

	if pauser.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the pauser")
	}

	paused := types.Paused{
		Paused: false,
	}

	k.SetPaused(ctx, paused)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUnpauseResponse{}, err
}
