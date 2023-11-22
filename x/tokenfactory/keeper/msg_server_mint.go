package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/noble-assets/noble/v4/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return k.Keeper.Mint(ctx, msg)
}

func (k Keeper) Mint(ctx sdk.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	minter, found := k.GetMinters(ctx, msg.From)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not a minter")
	}

	_, addressBz, err := bech32.DecodeAndConvert(msg.From)
	if err != nil {
		return nil, err
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minter address is blacklisted")
	}

	_, addressBz, err = bech32.DecodeAndConvert(msg.Address)
	if err != nil {
		return nil, err
	}

	_, found = k.GetBlacklisted(ctx, addressBz)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrMint, "receiver address is blacklisted")
	}

	mintingDenom := k.GetMintingDenom(ctx)

	if msg.Amount.Denom != mintingDenom.Denom {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting denom is incorrect")
	}

	if minter.Allowance.IsLT(msg.Amount) {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting amount is greater than the allowance")
	}

	paused := k.GetPaused(ctx)

	if paused.Paused {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting is paused")
	}

	minter.Allowance = minter.Allowance.Sub(msg.Amount)

	k.SetMinters(ctx, minter)

	amount := sdk.NewCoins(msg.Amount)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrMint, err.Error())
	}

	receiver, _ := sdk.AccAddressFromBech32(msg.Address)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiver, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrSendCoinsToAccount, err.Error())
	}

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgMintResponse{}, err
}
