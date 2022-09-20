package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ConfigureMinter(goCtx context.Context, msg *types.MsgConfigureMinter) (*types.MsgConfigureMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	mintingDenom, _ := k.GetMintingDenom(ctx)

	if msg.Allowance.Denom != mintingDenom.Denom {
		return nil, sdkerrors.Wrapf(types.ErrMint, "minting denom is incorrect")
	}

	minterController, found := k.GetMinterController(ctx, msg.From)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "minter controller not found")
	}

	if msg.From != minterController.Address {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not a controller of this minter")
	}

	minter, found := k.GetMinters(ctx, msg.Address)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "minter not found")
	}

	// TODO: https://github.com/strangelove-ventures/noble/issues/4

	minter.Allowance = msg.Allowance

	k.SetMinters(ctx, minter)

	return &types.MsgConfigureMinterResponse{}, nil
}
