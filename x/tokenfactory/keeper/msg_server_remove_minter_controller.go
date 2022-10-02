package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RemoveMinterController(goCtx context.Context, msg *types.MsgRemoveMinterController) (*types.MsgRemoveMinterControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	masterMinter, found := k.GetMasterMinter(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "master minter is not set")
	}

	if msg.From != masterMinter.Address {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the master minter")
	}

	_, found = k.GetMinterController(ctx, msg.Controller)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "minter controller with a given address (%s) doesn't exist", msg.Controller)
	}

	k.DeleteMinterController(ctx, msg.Controller)

	return &types.MsgRemoveMinterControllerResponse{}, nil
}
