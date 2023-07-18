package keeper

import (
	"context"

	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) UpdateSignatureThreshold(goCtx context.Context, msg *types.MsgUpdateSignatureThreshold) (*types.MsgUpdateSignatureThresholdResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Amount == 0 {
		return nil, sdkerrors.Wrapf(types.ErrUpdateSignatureThreshold, "invalid signature threshold")
	}

	authority, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if authority.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "this message sender cannot update the authority")
	}

	existingSignatureThreshold, found := k.GetSignatureThreshold(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrUpdateSignatureThreshold, "SignatureThreshold is not set")
	}

	if msg.Amount == existingSignatureThreshold.Amount {
		return nil, sdkerrors.Wrapf(types.ErrUpdateSignatureThreshold, "signature threshold already set")
	}

	// new signature threshold cannot be greater than the number of stored public keys
	publicKeys := k.GetAllPublicKeys(ctx)
	if msg.Amount > uint32(len(publicKeys)) {
		return nil, sdkerrors.Wrapf(types.ErrUpdateSignatureThreshold, "new signature threshold is too high")
	}

	newSignatureThreshold := types.SignatureThreshold{
		Amount: msg.Amount,
	}

	k.SetSignatureThreshold(ctx, newSignatureThreshold)

	event := types.SignatureThresholdUpdated{
		OldSignatureThreshold: uint64(existingSignatureThreshold.Amount),
		NewSignatureThreshold: uint64(newSignatureThreshold.Amount),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgUpdateSignatureThresholdResponse{}, err
}
