package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func (k msgServer) Burn(goCtx context.Context, msg *types.MsgBurn) (*types.MsgBurnResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, found := k.GetMinters(ctx, msg.From)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not a minter")
	}

	_, pubBz, err := bech32.DecodeAndConvert(msg.From)
	if err != nil {
		return nil, err
	}

	_, found = k.GetBlacklisted(ctx, pubBz)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrBurn, "minter address is blacklisted")
	}

	mintingDenom := k.GetMintingDenom(ctx)

	if msg.Amount.Denom != mintingDenom.Denom {
		return nil, sdkerrors.Wrapf(types.ErrMint, "burning denom is incorrect")
	}

	paused := k.GetPaused(ctx)

	if paused.Paused {
		return nil, sdkerrors.Wrapf(types.ErrBurn, "burning is paused")
	}

	minterAddress, _ := sdk.AccAddressFromBech32(msg.From)

	amount := sdk.NewCoins(msg.Amount)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, minterAddress, types.ModuleName, amount)
	if err != nil {
		return nil, err
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrBurn, err.Error())
	}

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgBurnResponse{}, err
}
