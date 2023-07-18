package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) UpdateMaxMessageBodySize(goCtx context.Context, msg *types.MsgUpdateMaxMessageBodySize) (*types.MsgUpdateMaxMessageBodySizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot update the max message body size")
	}

	newMaxMessageBodySize := types.MaxMessageBodySize{
		Amount: uint64(msg.Size_),
	}
	k.SetMaxMessageBodySize(ctx, newMaxMessageBodySize)

	event := types.MaxMessageBodySizeUpdated{
		NewMaxMessageBodySize: uint64(msg.Size_),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgUpdateMaxMessageBodySizeResponse{}, err
}
