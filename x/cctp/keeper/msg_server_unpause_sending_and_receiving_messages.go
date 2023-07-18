package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) UnpauseSendingAndReceivingMessages(goCtx context.Context, msg *types.MsgUnpauseSendingAndReceivingMessages) (*types.MsgUnpauseSendingAndReceivingMessagesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot unpause sending and receiving messages")
	}

	paused := types.SendingAndReceivingMessagesPaused{
		Paused: false,
	}
	k.SetSendingAndReceivingMessagesPaused(ctx, paused)

	event := types.UnpauseSendingAndReceiving{}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgUnpauseSendingAndReceivingMessagesResponse{}, err
}
