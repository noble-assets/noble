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
		return nil, sdkerrors.Wrapf(types.ErrUnauthorized, "This message sender cannot remove public keys")
	}

	_, found = k.GetAttester(ctx, string(msg.Attester))
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrRemovePublicKey, "Public Key not found in store, cannot delete")
	}

	// disallow removing public key if there is only 1 active public key
	storedPublicKeys := k.GetAllAttesters(ctx)
	if len(storedPublicKeys) == 1 {
		return nil, sdkerrors.Wrap(types.ErrRemovePublicKey, "Cannot remove public key if there only one left")
	}

	// disallow removing public key if it causes the n in m/n multisig to fall below m (threshold # of signers)
	signatureThreshold, found := k.GetSignatureThreshold(ctx)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrRemovePublicKey, "signature threshold not set")
	}

	if uint32(len(storedPublicKeys)) <= signatureThreshold.Amount-1 {
		return nil, sdkerrors.Wrap(types.ErrRemovePublicKey, "signature threshold is too low")
	}

	k.DeleteAttester(ctx, string(msg.Attester))

	event := types.AttesterDisabled{
		Attester: string(msg.Attester),
	}
	err := ctx.EventManager().EmitTypedEvent(&event)

	return &types.MsgDisableAttesterResponse{}, err
}
