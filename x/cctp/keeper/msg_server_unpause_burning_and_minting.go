package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) UnpauseBurningAndMinting(goCtx context.Context, msg *types.MsgUnpauseBurningAndMinting) (*types.MsgUnpauseBurningAndMintingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot unpause burning and minting")
	}

	paused := types.SendingAndReceivingMessagesPaused{
		Paused: false,
	}
	k.SetSendingAndReceivingMessagesPaused(ctx, paused)

	event := types.UnpauseBurningAndMinting{}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgUnpauseBurningAndMintingResponse{}, err
}
