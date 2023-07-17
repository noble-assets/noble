package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) EnableAttester(goCtx context.Context, msg *types.MsgEnableAttester) (*types.MsgEnableAttesterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot enable attesters")
	}

	_, found = k.GetAttester(ctx, string(msg.Attester))
	if found {
		return nil, sdkerrors.Wrapf(types.ErrAttesterAlreadyFound, "Attester already exists in store")
	}

	newAttester := types.Attester{
		Attester: string(msg.Attester),
	}
	k.SetAttester(ctx, newAttester)

	event := types.AttesterEnabled{
		Attester: string(msg.Attester),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgEnableAttesterResponse{}, err
}
