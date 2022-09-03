package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ChangeAdmin(goCtx context.Context, msg *types.MsgChangeAdmin) (*types.MsgChangeAdminResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin, found := k.GetAdmin(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "admin isn't set")
	}

	if admin.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the admin")
	}

	admin.Address = msg.Address

	k.SetAdmin(ctx, admin)

	return &types.MsgChangeAdminResponse{}, nil
}
