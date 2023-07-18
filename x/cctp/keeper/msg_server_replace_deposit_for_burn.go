package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) ReplaceDepositForBurn(goCtx context.Context, msg *types.MsgReplaceDepositForBurn) (*types.MsgReplaceDepositForBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	paused, found := k.GetSendingAndReceivingMessagesPaused(ctx)
	if found && paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "sending and receiving messages are paused")
	}

	// verify and parse original originalMessage
	if len(msg.OriginalMessage) < messageBodyIndex {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "invalid message: too short")
	}
	originalMessage := parseIntoMessage(msg.OriginalMessage)

	// verify and parse BurnMessage
	if len(originalMessage.MessageBody) != burnMessageLength {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "burn message body is not the correct length")
	}
	burnMessage := parseIntoBurnMessage(originalMessage.MessageBody)

	// validate originalMessage sender is the same as this message sender
	if msg.From != string(originalMessage.Sender) {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "Sender not permitted to use nonce")
	}

	// validate new mint recipient
	if len(msg.NewMintRecipient) == 0 {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "Mint recipient must be nonzero")
	}

	newMessageBody := types.BurnMessage{
		Version:       burnMessage.Version,
		BurnToken:     burnMessage.BurnToken,
		MintRecipient: msg.NewMintRecipient,
		Amount:        burnMessage.Amount,
		MessageSender: burnMessage.MessageSender,
	}

	newMessageBodyBytes := parseBurnMessageIntoBytes(newMessageBody)

	replaceMessage := types.MsgReplaceMessage{
		From:                 msg.From,
		OriginalMessage:      msg.OriginalMessage,
		OriginalAttestation:  msg.OriginalAttestation,
		NewMessageBody:       newMessageBodyBytes,
		NewDestinationCaller: msg.NewDestinationCaller,
	}
	_, err := k.ReplaceMessage(goCtx, &replaceMessage)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "error calling replace message")
	}

	event := types.DepositForBurn{
		Nonce:                     originalMessage.Nonce,
		BurnToken:                 string(burnMessage.BurnToken),
		Amount:                    burnMessage.Amount,
		Depositor:                 msg.From,
		MintRecipient:             msg.NewMintRecipient,
		DestinationDomain:         originalMessage.DestinationDomain,
		DestinationTokenMessenger: originalMessage.Recipient,
		DestinationCaller:         msg.NewDestinationCaller,
	}
	err = ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgReplaceDepositForBurnResponse{}, err
}
