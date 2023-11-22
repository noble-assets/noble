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

func (k msgServer) UpdateOwner(goCtx context.Context, msg *types.MsgUpdateOwner) (*types.MsgUpdateOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetOwner(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "owner is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the owner")
	}

	// ensure that the specified address is not already assigned to a privileged role
	err := k.ValidatePrivileges(ctx, msg.Address)
	if err != nil {
		return nil, err
	}

	owner.Address = msg.Address

	k.SetPendingOwner(ctx, owner)

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdateOwnerResponse{}, err
}
