package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateAuthority(goCtx context.Context, msg *types.MsgUpdateAuthority) (*types.MsgUpdateAuthorityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	currentAuthority, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if currentAuthority.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot update the authority")
	}

	newAuthority := types.Authority{
		Address: msg.NewAuthority,
	}
	k.SetAuthority(ctx, newAuthority)

	event := types.AuthorityUpdated{
		PreviousAuthority: currentAuthority.Address,
		NewAuthority:      newAuthority.Address,
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgUpdateAuthorityResponse{}, err
}
