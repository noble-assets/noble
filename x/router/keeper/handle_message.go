package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/strangelove-ventures/noble/x/router/types"
)

func (k Keeper) HandleMessage(ctx sdk.Context, msg []byte) error {

	// parse outer message
	outerMessage, err := DecodeMessage(msg)
	if err != nil {
		return err
	}

	// parse internal message into IBCForward
	if ibcForward, err := DecodeIBCForward(outerMessage.MessageBody); err == nil {
		if storedForward, ok := k.GetIBCForward(ctx, outerMessage.SourceDomain, string(outerMessage.Sender), outerMessage.Nonce); ok {
			if storedForward.AckError {
				if existingMint, ok := k.GetMint(ctx, outerMessage.SourceDomain, string(outerMessage.Sender), outerMessage.Nonce); ok {
					return k.ForwardPacket(ctx, ibcForward, existingMint)
				}
				panic("unexpected state")
			}

			return sdkerrors.Wrapf(types.ErrHandleMessage, "previous operation still in progress")
		}
		// this is the first time we are seeing this forward info -> store it.
		k.SetIBCForward(ctx, types.StoreIBCForwardMetadata{
			SourceDomainSender: string(outerMessage.Sender),
			Nonce:              outerMessage.Nonce,
			Metadata:           &ibcForward,
		})
		if existingMint, ok := k.GetMint(ctx, outerMessage.SourceDomain, string(outerMessage.Sender), outerMessage.Nonce); ok {
			return k.ForwardPacket(ctx, ibcForward, existingMint)
		}
		return nil
	}

	// try to parse internal message into burn (representing a remote burn -> local mint)
	if burnMessage, err := DecodeBurnMessage(outerMessage.MessageBody); err == nil {
		// look up corresponding mint token from cctp
		tokenPair, found := k.cctpKeeper.GetTokenPair(ctx, outerMessage.SourceDomain, string(burnMessage.BurnToken))
		if !found {
			return sdkerrors.Wrapf(types.ErrHandleMessage, "unable to find local token denom for this burn")
		}

		// message is a Mint
		mint := types.Mint{
			SourceDomainSender: string(outerMessage.Sender),
			Nonce:              outerMessage.Nonce,
			Amount: &sdk.Coin{
				Denom:  tokenPair.LocalToken,
				Amount: sdk.NewIntFromBigInt(&burnMessage.Amount),
			},
			DestinationDomain: strconv.Itoa(int(outerMessage.DestinationDomain)),
			MintRecipient:     string(burnMessage.MintRecipient),
		}
		k.SetMint(ctx, mint)
		if existingIBCForward, found := k.GetIBCForward(ctx, outerMessage.SourceDomain, string(burnMessage.MessageSender), outerMessage.Nonce); found {
			return k.ForwardPacket(ctx, *existingIBCForward.Metadata, mint)
		}
	}

	return nil
}

func (k Keeper) ForwardPacket(ctx sdk.Context, ibcForward types.IBCForwardMetadata, mint types.Mint) error {
	timeout := ibcForward.TimeoutInNanoseconds
	if timeout == 0 {
		timeout = transfertypes.DefaultRelativePacketTimeoutTimestamp
	}

	transfer := &transfertypes.MsgTransfer{
		SourcePort:    ibcForward.Port,
		SourceChannel: ibcForward.Channel,
		Token:         *mint.Amount,
		Sender:        mint.MintRecipient,
		Receiver:      ibcForward.DestinationReceiver,
		TimeoutHeight: clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 0,
		},
		TimeoutTimestamp: uint64(ctx.BlockTime().UnixNano()) + timeout,
		Memo:             ibcForward.Memo,
	}

	res, err := k.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transfer)
	if err != nil {
		return err
	}

	inFlightPacket := types.InFlightPacket{
		SourceDomainSender: mint.SourceDomainSender,
		Nonce:              mint.Nonce,
		ChannelId:          ibcForward.Channel,
		PortId:             ibcForward.Port,
		Sequence:           res.Sequence,
	}

	k.SetInFlightPacket(ctx, inFlightPacket)

	return nil
}
