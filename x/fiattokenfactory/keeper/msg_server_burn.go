package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func (k msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := k.GetMinters(ctx, msg.From)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrBurn, "%v: you are not a minter", types.ErrUnauthorized)
	}

	_, addressBz, err := bech32.DecodeAndConvert(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, sdkerrors.Wrap(types.ErrBurn, "minter address is blacklisted")
	}

	mintingDenom := k.GetMintingDenom(ctx)

	if msg.Amount.Denom != mintingDenom.Denom {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning denom is incorrect")
	}

	paused := k.GetPaused(ctx)

	if paused.Paused {
		return nil, sdkerrors.Wrap(types.ErrBurn, "burning is paused")
	}

	minterAddress, err := sdk.AccAddressFromBech32(msg.From)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	amount := sdk.NewCoins(msg.Amount)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, minterAddress, types.ModuleName, amount)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgBurnResponse{}, err
}
