package keeper

import (
	"context"

	"noble/x/tokenfactory/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) ConfigureMinter(goCtx context.Context, msg *types.MsgConfigureMinter) (*types.MsgConfigureMinterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	masterMinter, _ := k.GetMasterMinter(ctx)

	minterController, minterControllerFound := k.GetMinterController(ctx, msg.Address)

	notMasterMinter := msg.From != masterMinter.Address
	notMinterController := minterControllerFound && msg.From != minterController.Address

	if notMasterMinter || notMinterController {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "you are not the master minter or a minter controller")
	}

	// TODO: https://github.com/strangelove-ventures/noble/issues/4

	minter := types.Minters{
		Address:   msg.Address,
		Allowance: msg.Allowance,
	}

	k.SetMinters(ctx, minter)

	return &types.MsgConfigureMinterResponse{}, nil
}
