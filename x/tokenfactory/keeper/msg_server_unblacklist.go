package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

func (k msgServer) Unblacklist(goCtx context.Context, msg *types.MsgUnblacklist) (*types.MsgUnblacklistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgUnblacklistResponse{}, nil
}
