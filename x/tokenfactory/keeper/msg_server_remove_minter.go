package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RemoveMinter(goCtx context.Context, msg *types.MsgRemoveMinter) (*types.MsgRemoveMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetMasterMinter(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "master minter isn't set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the master minter")
	}

	minter, found := k.GetMinters(ctx, msg.Address)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "a minter with a given address doesn't exist")
	}

	k.RemoveMinters(ctx, minter.Address)

	return &types.MsgRemoveMinterResponse{}, nil
}
