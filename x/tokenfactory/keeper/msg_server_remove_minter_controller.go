package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

func (k msgServer) RemoveMinterController(goCtx context.Context, msg *types.MsgRemoveMinterController) (*types.MsgRemoveMinterControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRemoveMinterControllerResponse{}, nil
}
