package keeper

import (
	"context"
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) UpdatePerMessageBurnLimit(goCtx context.Context, msg *types.MsgUpdatePerMessageBurnLimit) (*types.MsgUpdatePerMessageBurnLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot update the per message burn limit")
	}

	newPerMessageBurnLimit := types.PerMessageBurnLimit{
		Amount: msg.PerMessageBurnLimit,
	}
	k.SetPerMessageBurnLimit(ctx, newPerMessageBurnLimit)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdatePerMessageBurnLimitResponse{}, err
}
