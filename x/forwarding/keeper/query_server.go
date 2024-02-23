package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/noble-assets/noble/v4/x/forwarding/types"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) Address(goCtx context.Context, req *types.QueryAddress) (*types.QueryAddressResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	address := types.GenerateAddress(req.Channel, req.Recipient)

	exists := false
	if k.authKeeper.HasAccount(ctx, address) {
		account := k.authKeeper.GetAccount(ctx, address)
		_, exists = account.(*types.ForwardingAccount)
	}

	return &types.QueryAddressResponse{
		Address: address.String(),
		Exists:  exists,
	}, nil
}

func (k *Keeper) StatsByChannel(goCtx context.Context, req *types.QueryStatsByChannel) (*types.QueryStatsByChannelResponse, error) {
	if req == nil {
		return nil, errors.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryStatsByChannelResponse{
		NumOfAccounts:  k.GetNumOfAccounts(ctx, req.Channel),
		NumOfForwards:  k.GetNumOfForwards(ctx, req.Channel),
		TotalForwarded: k.GetTotalForwarded(ctx, req.Channel),
	}, nil
}
