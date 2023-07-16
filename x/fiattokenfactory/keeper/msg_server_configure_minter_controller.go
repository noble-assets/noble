package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"

	sdkerrors "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ConfigureMinterController(goCtx context.Context, msg *types.MsgConfigureMinterController) (*types.MsgConfigureMinterControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	masterMinter, found := k.GetMasterMinter(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "master minter is not set")
	}

	if masterMinter.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the master minter")
	}

	controller := types.MinterController{
		Minter:     msg.Minter,
		Controller: msg.Controller,
	}

	k.SetMinterController(ctx, controller)

	return &types.MsgConfigureMinterControllerResponse{}, nil
}
