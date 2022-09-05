package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Unblacklist(goCtx context.Context, msg *types.MsgUnblacklist) (*types.MsgUnblacklistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	blacklister, found := k.GetBlacklister(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "blacklister is not set")
	}

	if blacklister.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the blacklister")
	}

	blacklisted, found := k.GetBlacklisted(ctx, msg.Address)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "a given address is not blacklisted")
	}

	k.RemoveBlacklisted(ctx, blacklisted.Address)

	return &types.MsgUnblacklistResponse{}, nil
}
