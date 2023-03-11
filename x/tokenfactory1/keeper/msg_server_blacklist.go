package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/strangelove-ventures/noble/x/tokenfactory1/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Blacklist(goCtx context.Context, msg *types.MsgBlacklist) (*types.MsgBlacklistResponse, error) {
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

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, types.ErrUserBlacklisted
	}

	blacklisted := types.Blacklisted{
		AddressBz: addressBz,
	}

	k.SetBlacklisted(ctx, blacklisted)

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgBlacklistResponse{}, err
}
