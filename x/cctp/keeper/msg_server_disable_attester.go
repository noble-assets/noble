package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func (k msgServer) DisableAttester(goCtx context.Context, msg *types.MsgDisableAttester) (*types.MsgDisableAttesterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, found := k.GetAuthority(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrAuthorityNotSet, "Authority is not set")
	}

	if owner.Address != msg.From {
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot disable an attester")
	}

	_, found = k.GetAttester(ctx, string(msg.Attester))
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrDisableAttester, "Attester not found in store, cannot delete")
	}

	// disallow disabled attester if there is only 1 active attester
	storedAttesters := k.GetAllAttesters(ctx)
	if len(storedAttesters) == 1 {
		return nil, sdkerrors.Wrap(types.ErrDisableAttester, "Cannot disable an attester if there is only one left")
	}

	// disallow disabling attester if it causes the n in m/n multisig to fall below m (threshold # of signers)
	signatureThreshold, found := k.GetSignatureThreshold(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrDisableAttester, "signature threshold not set")
	}

	if uint32(len(storedAttesters)) <= signatureThreshold.Amount-1 {
		return nil, sdkerrors.Wrap(types.ErrDisableAttester, "signature threshold is too low")
	}

	k.DeleteAttester(ctx, string(msg.Attester))

	event := types.AttesterDisabled{
		Attester: string(msg.Attester),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgDisableAttesterResponse{}, err
}
