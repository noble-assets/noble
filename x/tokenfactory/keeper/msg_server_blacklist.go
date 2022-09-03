package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

func (k msgServer) Blacklist(goCtx context.Context, msg *types.MsgBlacklist) (*types.MsgBlacklistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgBlacklistResponse{}, nil
}
