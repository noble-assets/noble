package keeper

import (
	"bytes"
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) AddTokenMessenger(goCtx context.Context, msg *types.MsgAddTokenMessengerRequest) (*types.MsgAddTokenMessengerResponse, error) {
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
	if found {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "TokenMessenger already set")
	}

	newTokenMessenger := types.TokenMessenger{
		DomainId: msg.DomainId,
		Address:  msg.Address,
	}
	k.SetTokenMessenger(ctx, newTokenMessenger)

	event := types.RemoteTokenMessengerAdded{
		Domain:         msg.DomainId,
		TokenMessenger: []byte(msg.Address),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgAddTokenMessengerResponse{}, err
}
