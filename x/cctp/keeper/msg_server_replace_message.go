package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) ReplaceMessage(goCtx context.Context, msg *types.MsgReplaceMessage) (*types.MsgReplaceMessageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	paused, found := k.GetSendingAndReceivingMessagesPaused(ctx)
	if found && paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "sending and receiving messages are paused")
	}

	// Validate each signature in the attestation
	publicKeys := k.GetAllPublicKeys(ctx)
	if len(publicKeys) == 0 {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "no public keys found")
	}

	signatureThreshold, found := k.GetSignatureThreshold(ctx)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "signature threshold not found")
	}

	verified, err := verifyAttestationSignatures(msg.OriginalMessage, msg.OriginalAttestation, publicKeys, signatureThreshold.Amount)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "error during signature verification")
	}
	if !verified {
		return nil, sdkerrors.Wrapf(err, "unable to verify signatures")
	}

	// validate message format
	if len(msg.OriginalMessage) < messageBodyIndex {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "invalid message: too short")
	}
	originalMessage := parseIntoMessage(msg.OriginalMessage)

	// validate originalMessage sender is the same as this message sender
	if msg.From != string(originalMessage.Sender) {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "sender not permitted to use nonce")
	}

	// validate source domain
	if originalMessage.SourceDomain != nobleDomainId {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "message not originally sent from this domain")
	}

	sendMessage := types.MsgSendMessage{
		From:              msg.From,
		DestinationDomain: originalMessage.DestinationDomain,
		Recipient:         originalMessage.Recipient,
		MessageBody:       msg.NewMessageBody,
	}

	_, err = k.SendMessage(goCtx, &sendMessage)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "error during send message")
	}

	return &types.MsgReplaceMessageResponse{}, err
}
