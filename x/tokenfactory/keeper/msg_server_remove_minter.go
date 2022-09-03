package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

func (k msgServer) RemoveMinter(goCtx context.Context, msg *types.MsgRemoveMinter) (*types.MsgRemoveMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgRemoveMinterResponse{}, nil
}
