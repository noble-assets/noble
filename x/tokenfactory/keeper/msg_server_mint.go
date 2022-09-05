package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	minter, found := k.GetMinters(ctx, msg.From)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not a minter")
	}

	_, found = k.GetBlacklisted(ctx, msg.From)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minter address is blacklisted")
	}

	_, found = k.GetBlacklisted(ctx, msg.Address)
	if found {
		return nil, sdkerrors.Wrapf(types.ErrMint, "receiver address is blacklisted")
	}

	if minter.Allowance.IsLT(msg.Amount) {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting amount is greater than the allowance")
	}

	// TODO: add checking for paused

	minter.Allowance = minter.Allowance.Sub(msg.Amount)

	amount := sdk.NewCoins(msg.Amount)

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrMint, err.Error())
	}

	reciever, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(types.ErrParseAddress, err.Error())
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, reciever, amount); err != nil {
		return nil, sdkerrors.Wrap(types.ErrSendCoinsToAccount, err.Error())
	}

	return &types.MsgMintResponse{}, nil
}
