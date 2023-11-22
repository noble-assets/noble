package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/msg_server_accept_owner.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/msg_server_accept_owner.go

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) AcceptOwner(goCtx context.Context, msg *types.MsgAcceptOwner) (*types.MsgAcceptOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetPendingOwner(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "pending owner is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the pending owner")
	}

	k.SetOwner(ctx, owner)

	k.DeletePendingOwner(ctx)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgAcceptOwnerResponse{}, err
}
