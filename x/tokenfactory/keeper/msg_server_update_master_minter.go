package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateMasterMinter(goCtx context.Context, msg *types.MsgUpdateMasterMinter) (*types.MsgUpdateMasterMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetOwner(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "owner is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the owner")
	}

	masterMinter := types.MasterMinter{
		Address: msg.Address,
	}

	k.SetMasterMinter(ctx, masterMinter)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdateMasterMinterResponse{}, err
}
