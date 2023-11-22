package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/msg_server_remove_minter_controller.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/msg_server_remove_minter_controller.go

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) RemoveMinterController(goCtx context.Context, msg *types.MsgRemoveMinterController) (*types.MsgRemoveMinterControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	masterMinter, found := k.GetMasterMinter(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "master minter is not set")
	}

	if msg.From != masterMinter.Address {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the master minter")
	}

	_, found = k.GetMinterController(ctx, msg.Controller)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "minter controller with a given address (%s) doesn't exist", msg.Controller)
	}

	k.DeleteMinterController(ctx, msg.Controller)

	return &types.MsgRemoveMinterControllerResponse{}, nil
}
