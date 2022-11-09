package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

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

	blacklister := types.Blacklister{
		Address: msg.Address,
	}

	k.SetBlacklister(ctx, blacklister)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdateBlacklisterResponse{}, err
}
