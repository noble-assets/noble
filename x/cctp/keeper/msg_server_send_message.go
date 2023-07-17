package keeper

import (
	"bytes"
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) SendMessage(goCtx context.Context, msg *types.MsgSendMessage) (*types.MsgSendMessageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	paused, found := k.GetSendingAndReceivingMessagesPaused(ctx)
	if found && paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrSendMessage, "sending and receiving messages is paused")
	}

	// check if message body is too long, ignore if max length not found
	max, found := k.GetMaxMessageBodySize(ctx)
	if found && uint64(len(msg.MessageBody)) > max.Amount {
		return nil, sdkerrors.Wrap(types.ErrSendMessage, "message body exceeds max size")
	}

	emptyByteArr := make([]byte, len(msg.Recipient))
	if len(msg.Recipient) == 0 || bytes.Equal(msg.Recipient, emptyByteArr) {
		return nil, sdkerrors.Wrap(types.ErrSendMessage, "recipient must not be nonzero")
	}

	event := types.MessageSent{
		Message: msg.MessageBody,
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgSendMessageResponse{}, err
}
