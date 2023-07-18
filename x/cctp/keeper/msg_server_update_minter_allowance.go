package keeper

import (
	"context"
	"strings"

	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateMinterAllowance(goCtx context.Context, msg *types.MsgUpdateMinterAllowance) (*types.MsgUpdateMinterAllowanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if authority.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot update the minter allowance")
	}

	_, found = k.GetMinterAllowance(ctx, strings.ToLower(msg.Denom))
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrMinterAllowanceNotFound, "MinterAllowance is not set")
	}

	allowance := types.MinterAllowances{
		Denom:  msg.Denom,
		Amount: msg.Amount,
	}

	k.SetMinterAllowance(ctx, allowance)

	err := ctx.EventManager().EmitTypedEvent(msg)

	return &types.MsgUpdateMinterAllowanceResponse{}, err
}
