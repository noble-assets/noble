package keeper

import (
	"context"

<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Pause(goCtx context.Context, msg *types.MsgPause) (*types.MsgPauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pauser, found := k.GetPauser(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "pauser is not set")
	}

	if pauser.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the pauser")
	}

	paused := types.Paused{
		Paused: true,
	}

	k.SetPaused(ctx, paused)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgPauseResponse{}, err
}
