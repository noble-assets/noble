package keeper

import (
	"bytes"
	"context"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
)

func (k msgServer) DepositForBurn(goCtx context.Context, msg *types.MsgDepositForBurn) (*types.MsgDepositForBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Amount <= 0 {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "amount must be nonzero")
	}

	emptyByteArr := make([]byte, 32)
	if bytes.Equal(msg.MintRecipient, emptyByteArr) {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "mint recipient must be nonzero")
	}

	// hardcoded lookup
	destinationTokenMessenger := []byte(TokenMessengerMap[msg.DestinationDomain])
	if len(destinationTokenMessenger) == 0 {
		return nil, sdkerrors.Wrap(types.ErrDepositForBurn, "unable to look up destination token messenger")
	}

	denom := k.fiattokenfactory.GetMintingDenom(ctx)
	if denom.Denom != strings.ToLower(msg.BurnToken) {
		return nil, sdkerrors.Wrapf(types.ErrBurn, "burning denom: %s is not supported", msg.BurnToken)
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
	var fiatBurnMsg = fiattokenfactorytypes.MsgBurn{
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
		Version:       MessageBodyVersion,
		BurnToken:     []byte(msg.BurnToken),
		MintRecipient: msg.MintRecipient,
		Amount:        uint64(msg.Amount),
		MessageSender: []byte(msg.From),
	}
	burnMessageBytes := burnMessage.Bytes()

	sendMessage := types.MsgSendMessage{
		From:              msg.From,
		DestinationDomain: msg.DestinationDomain,
		Recipient:         burnMessage.MintRecipient,
		MessageBody:       burnMessageBytes,
	}

	// send message
	_, err = k.SendMessage(goCtx, &sendMessage)
	if err != nil {
		return nil, err
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
		DestinationTokenMessenger: destinationTokenMessenger,
		DestinationCaller:         nil,
	}
	err = ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgDepositForBurnResponse{Nonce: nonceReserved.Nonce}, err
}
