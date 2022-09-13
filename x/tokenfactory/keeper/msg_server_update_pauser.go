package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdatePauser(goCtx context.Context, msg *types.MsgUpdatePauser) (*types.MsgUpdatePauserResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetOwner(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "owner is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the owner")
	}

	pauser := types.Pauser{
		Address: msg.Address,
	}

	k.SetPauser(ctx, pauser)

	return &types.MsgUpdatePauserResponse{}, nil
}
