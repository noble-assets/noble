package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Unblacklist(goCtx context.Context, msg *types.MsgUnblacklist) (*types.MsgUnblacklistResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	blacklister, found := k.GetBlacklister(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "blacklister is not set")
	}

	if blacklister.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the blacklister")
	}

	_, addressBz, err := bech32.DecodeAndConvert(msg.Address)
	if err != nil {
		return nil, err
	}

	blacklisted, found := k.GetBlacklisted(ctx, addressBz)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "the specified address is not blacklisted")
	}

	k.RemoveBlacklisted(ctx, blacklisted.AddressBz)

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUnblacklistResponse{}, err
}
