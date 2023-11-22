package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/msg_server_configure_minter_controller.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/msg_server_configure_minter_controller.go

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
