package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) SoftwareUpgrade(goCtx context.Context, msg *types.MsgSoftwareUpgrade) (*types.MsgSoftwareUpgradeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.upgradeKeeper.ScheduleUpgrade(ctx, *msg.Plan)

	return &types.MsgSoftwareUpgradeResponse{}, nil
}
