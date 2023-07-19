package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) LinkTokenPair(goCtx context.Context, msg *types.MsgLinkTokenPair) (*types.MsgLinkTokenPairResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot link token pairs")
	}

	// check whether there already exists a mapping for this remote domain/token
	_, found = k.GetTokenPair(ctx, msg.RemoteDomain, msg.RemoteToken)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrTokenPairAlreadyFound, "Token pair for this remote domain, token already exists in store")
	}

	newTokenPair := types.TokenPairs{
		RemoteDomain: msg.RemoteDomain,
		RemoteToken:  msg.RemoteToken,
		LocalToken:   msg.LocalToken,
	}

	k.SetTokenPair(ctx, newTokenPair)

	event := types.TokenPairLinked{
		LocalToken:   newTokenPair.LocalToken,
		RemoteDomain: newTokenPair.RemoteDomain,
		RemoteToken:  newTokenPair.RemoteToken,
	}
	err := ctx.EventManager().EmitTypedEvent(&event)
	return &types.MsgLinkTokenPairResponse{}, err
}
