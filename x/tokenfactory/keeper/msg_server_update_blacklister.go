package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"noble/x/tokenfactory/types"
)

func (k msgServer) UpdateBlacklister(goCtx context.Context, msg *types.MsgUpdateBlacklister) (*types.MsgUpdateBlacklisterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgUpdateBlacklisterResponse{}, nil
}
