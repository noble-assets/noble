package keeper

import (
	"context"

<<<<<<< HEAD:x/fiattokenfactory/keeper/msg_server_update_master_minter.go
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283)):x/stabletokenfactory/keeper/msg_server_update_master_minter.go

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateMasterMinter(goCtx context.Context, msg *types.MsgUpdateMasterMinter) (*types.MsgUpdateMasterMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetOwner(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUserNotFound, "owner is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the owner")
	}

	// ensure that the specified address is not already assigned to a privileged role
	err := k.ValidatePrivileges(ctx, msg.Address)
	if err != nil {
		return nil, err
	}

	masterMinter := types.MasterMinter{
		Address: msg.Address,
	}

	k.SetMasterMinter(ctx, masterMinter)

	err = ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdateMasterMinterResponse{}, err
}
