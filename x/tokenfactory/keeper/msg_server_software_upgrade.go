package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SoftwareUpgrade(goCtx context.Context, msg *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	admin, found := k.GetAdmin(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "admin is not set")
	}

	if admin.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the admin")
	}

	k.upgradeKeeper.ScheduleUpgrade(ctx, *msg.Plan)

	return &types.MsgSoftwareUpgradeResponse{}, nil
}
