package keeper

import (
	"bytes"
	"context"
	"github.com/strangelove-ventures/noble/x/cctp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) RemoveTokenMessenger(goCtx context.Context, msg *types.MsgAddTokenMessenger) (*types.MsgAddTokenMessengerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot add token messengers")
	}

	emptyByteArr := make([]byte, cctp.Bytes32Len)
	if bytes.Equal([]byte(msg.Address), emptyByteArr) {
		return nil, sdkerrors.Wrapf(types.ErrMalformedField, "token messenger address cannot be 0")
	}

	_, found = k.GetTokenMessenger(ctx, msg.DomainId)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "TokenMessenger not found for this domain id")
	}

	k.DeleteTokenMessenger(ctx, msg.DomainId)

	event := types.RemoteTokenMessengerRemoved{
		Domain:         msg.DomainId,
		TokenMessenger: []byte(msg.Address),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgAddTokenMessengerResponse{}, err
}
