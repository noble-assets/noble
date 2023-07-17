package keeper

import (
	"bytes"
	"context"
	"github.com/strangelove-ventures/noble/x/cctp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
)

func (k msgServer) DepositForBurnWithCaller(goCtx context.Context, msg *types.MsgDepositForBurnWithCaller) (*types.MsgDepositForBurnWithCallerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Destination caller must be nonzero. To allow any destination caller, use DepositForBurn().
	emptyByteArr := make([]byte, cctp.Bytes32Len)
	if len(msg.DestinationCaller) != cctp.Bytes32Len || bytes.Equal(msg.DestinationCaller, emptyByteArr) {
		return nil, sdkerrors.Wrap(types.ErrInvalidDestinationCaller, "destination caller must be nonzero")
	}

	if msg.Amount <= 0 {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "amount must be nonzero")
	}

	emptyByteArr = make([]byte, cctp.Bytes32Len)
	if bytes.Equal(msg.MintRecipient, emptyByteArr) {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "mint recipient must be nonzero")
	}

	destinationTokenMessenger, found := k.GetTokenMessenger(ctx, msg.DestinationDomain)
	if !found {
		return nil, sdkerrors.Wrap(types.ErrBurn, "failed to look up destination token messenger")
	}

	// verify token denom is supported for burning
	_, found = k.router.GetDenom(ctx, strings.ToLower(msg.BurnToken))
	if !found {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning this denom is not supported")
	}

	// check if burning/minting is paused
	paused, found := k.GetBurningAndMintingPaused(ctx)
	if found && paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning and minting are paused")
	}

	// check if amount is greater than configured PerMessageBurnLimit
	perMessageBurnLimit, found := k.GetPerMessageBurnLimit(ctx)
	if found {
		if uint64(msg.Amount) > perMessageBurnLimit.Amount {
			return nil, sdkerrors.Wrap(types.ErrBurn, "cannot burn more than the maximum per message burn limit")
		}
	}

	// burn coins
	fiatBurnMsg := fiattokenfactorytypes.MsgBurn{
		From: msg.From,
		Amount: sdk.Coin{
			Denom:  msg.BurnToken,
			Amount: sdk.NewInt(int64(msg.Amount)),
		},
	}
	_, err := k.fiattokenfactory.Burn(goCtx, &fiatBurnMsg)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "error during burn ")
	}

	// get burn message into bytes
	burnMessage := types.BurnMessage{
		Version:       cctp.MessageBodyVersion,
		BurnToken:     []byte(msg.BurnToken),
		MintRecipient: msg.MintRecipient,
		Amount:        uint64(msg.Amount),
		MessageSender: []byte(msg.From),
	}
	burnMessageBytes := parseBurnMessageIntoBytes(burnMessage)

	sendMessage := types.MsgSendMessageWithCaller{
		From:              msg.From,
		DestinationDomain: msg.DestinationDomain,
		Recipient:         burnMessage.MintRecipient,
		MessageBody:       burnMessageBytes,
		DestinationCaller: msg.DestinationCaller,
	}

	// send message
	_, err = k.SendMessageWithCaller(goCtx, &sendMessage)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "error during send message with caller")
	}

	// reserve and increment nonce
	nonceReserved, found := k.GetNonce(ctx)
	nextAvailableNonce := types.Nonce{
		Nonce: nonceReserved.Nonce + 1,
	}
	k.SetNonce(ctx, nextAvailableNonce)

	event := types.DepositForBurn{
		Nonce:                     nonceReserved.Nonce,
		BurnToken:                 msg.BurnToken,
		Amount:                    uint64(msg.Amount),
		Depositor:                 msg.From,
		MintRecipient:             msg.MintRecipient,
		DestinationDomain:         msg.DestinationDomain,
		DestinationTokenMessenger: []byte(destinationTokenMessenger.Address),
		DestinationCaller:         msg.DestinationCaller,
	}
	err = ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgDepositForBurnWithCallerResponse{Nonce: nonceReserved.Nonce}, nil
}
