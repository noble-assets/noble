package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Mint(goCtx context.Context, msg *types.MsgMint) (*types.MsgMintResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO:
	// Fails if:
	// - Message creator is not a minter or is blacklisted
	// - `address` is blacklisted
	// - `isPaused` is `true`
	// - `amount` > minter’s allowance
	// - `amount` has a wrong denom

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

	// Add events

	return &types.MsgMintResponse{}, nil
}
