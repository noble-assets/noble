package keeper

import (
	"context"

<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateBlacklister(goCtx context.Context, msg *types.MsgUpdateBlacklister) (*types.MsgUpdateBlacklisterResponse, error) {
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

	blacklister := types.Blacklister{
		Address: msg.Address,
	}

	k.SetBlacklister(ctx, blacklister)

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdateBlacklisterResponse{}, err
}
