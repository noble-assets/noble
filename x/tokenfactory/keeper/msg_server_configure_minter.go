package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

func (k msgServer) ConfigureMinter(goCtx context.Context, msg *types.MsgConfigureMinter) (*types.MsgConfigureMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgConfigureMinterResponse{}, nil
}
