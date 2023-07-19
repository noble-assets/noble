package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
)

func (k msgServer) ReceiveMessage(goCtx context.Context, msg *types.MsgReceiveMessage) (*types.MsgReceiveMessageResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	paused, found := k.GetSendingAndReceivingMessagesPaused(ctx)
	if found && paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrReceiveMessage, "sending and receiving messages are paused")
	}

	//// Validate each signature in the attestation
	publicKeys := k.GetAllPublicKeys(ctx)
	if len(publicKeys) == 0 {
		return nil, sdkerrors.Wrap(types.ErrReplaceMessage, "no public keys found")
	}

	signatureThreshold, found := k.GetSignatureThreshold(ctx)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrReceiveMessage, "signature threshold not found")
	}

	verified, err := VerifyAttestationSignatures(msg.Message, msg.Attestation, publicKeys, signatureThreshold.Amount)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "error during signature verification")
	}
	if !verified {
		return nil, sdkerrors.Wrapf(types.ErrReplaceMessage, "unable to verify signatures")
	}

	// verify and parse message
	message := new(types.Message)
	if err := message.UnmarshalBytes(msg.Message); err != nil {
		return nil, sdkerrors.Wrap(types.ErrReceiveMessage, err.Error())
	}

	// validate correct domain
	if message.DestinationDomain != nobleDomainId {
		return nil, sdkerrors.Wrapf(types.ErrReceiveMessage, "incorrect destination domain: %d", message.DestinationDomain)
	}

	// validate destination caller
	emptyByteArr := make([]byte, 32)
	if !bytes.Equal(message.DestinationCaller, emptyByteArr) && string(message.DestinationCaller) != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrReceiveMessage, "incorrect destination caller: %s, sender: %s", message.DestinationCaller, msg.From)
	}

	// validate version
	if message.Version != nobleVersion {
		return nil, sdkerrors.Wrapf(types.ErrReceiveMessage, "invalid message version. expected: %d, found: %d", nobleVersion, message.Version)
	}

	sourceDomainBz := make([]byte, 4)
	binary.BigEndian.PutUint32(sourceDomainBz, message.SourceDomain)

	nonceBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBz, message.Nonce)

	// verify nonce has not been used
	usedNonceKey := UsedNonce{
		Nonce:        message.Nonce,
		SourceDomain: message.SourceDomain,
	}
	found = k.GetUsedNonce(ctx, usedNonceKey)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrReceiveMessage, "nonce already used")
	}

	// mark nonce as used
	k.SetUsedNonce(ctx, usedNonceKey)

	// verify and parse BurnMessage
	burnMessage := new(types.BurnMessage)
	if err := burnMessage.UnmarshalBytes(message.MessageBody); err == nil { // mint

		nonZeroIndex := 0
		for i, b := range burnMessage.BurnToken {
			if b != 0 {
				nonZeroIndex = i
				break
			}
		}

		unpadded := burnMessage.BurnToken[nonZeroIndex:]
		burnTokenHex := hex.EncodeToString(unpadded)

		// look up Noble mint token from corresponding source domain/token
		tokenPair, found := k.GetTokenPair(ctx, message.SourceDomain, burnTokenHex)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrReceiveMessage, "corresponding noble mint token not found for %s", burnTokenHex)
		}

		msgMint := fiattokenfactorytypes.MsgMint{
			From:    msg.From,
			Address: string(burnMessage.MintRecipient),
			Amount: sdk.Coin{
				Denom:  strings.ToLower(tokenPair.LocalToken),
				Amount: sdk.NewIntFromUint64(burnMessage.Amount),
			},
		}

		_, err = k.fiattokenfactory.Mint(goCtx, &msgMint)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "Error during minting")
		}

		mintEvent := types.MintAndWithdraw{
			MintRecipient: string(burnMessage.MintRecipient),
			Amount:        burnMessage.Amount,
			MintToken:     string(burnMessage.BurnToken),
		}
		err = ctx.EventManager().EmitTypedEvent(&mintEvent)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "Error emitting mint event")
		}
	}

	if err := k.router.HandleMessage(ctx, msg.Message); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrMint, "Error in handleMessage")
	}

	event := types.MessageReceived{
		Caller:       msg.From,
		SourceDomain: message.SourceDomain,
		Nonce:        message.Nonce,
		Sender:       message.Sender,
		MessageBody:  message.MessageBody,
	}
	err = ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgReceiveMessageResponse{Success: true}, err
}
