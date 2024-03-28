package keeper

import (
	"context"

	"github.com/noble-assets/noble/v5/x/tokenfactory/types"

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

	// check if controller has already been assigned to a minter and the minter has non-zero allowance
	mc, found := k.GetMinterController(ctx, msg.Controller)
	if found {
		m, f := k.GetMinters(ctx, mc.Minter)
		if f && mc.Minter != msg.Minter && !m.Allowance.IsZero() {
			return nil, sdkerrors.Wrapf(types.ErrConfigureController, "its assigned minter still has allowance")
		}
	}

	controller := types.MinterController{
		Minter:     msg.Minter,
		Controller: msg.Controller,
	}

	k.SetMinterController(ctx, controller)

	return &types.MsgConfigureMinterControllerResponse{}, nil
}
