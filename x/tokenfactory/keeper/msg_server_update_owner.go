package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateOwner(goCtx context.Context, msg *types.MsgUpdateOwner) (*types.MsgUpdateOwnerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message
	_ = ctx

	return &types.MsgUpdateOwnerResponse{}, nil
}
