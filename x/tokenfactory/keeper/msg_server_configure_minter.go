package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ConfigureMinter(goCtx context.Context, msg *types.MsgConfigureMinter) (*types.MsgConfigureMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	masterMinter, found := k.GetMasterMinter(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "master minter is not set")
	}

	if masterMinter.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the master minter")
	}

	// TODO: https://github.com/strangelove-ventures/noble/issues/4

	minter := types.Minters{
		Address:   msg.Address,
		Allowance: msg.Allowance,
	}

	k.SetMinters(ctx, minter)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgConfigureMinterResponse{}, err
}
