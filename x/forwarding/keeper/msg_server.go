package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/noble-assets/noble/v5/x/forwarding/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := types.GenerateAddress(msg.Channel, msg.Recipient)

	channel, found := k.channelKeeper.GetChannel(ctx, transfertypes.PortID, msg.Channel)
	if !found {
		return nil, fmt.Errorf("channel does not exist: %s", msg.Channel)
	}
	if channel.State != channeltypes.OPEN {
		return nil, fmt.Errorf("channel is not open: %s, %s", msg.Channel, channel.State)
	}

	if k.authKeeper.HasAccount(ctx, address) {
		rawAccount := k.authKeeper.GetAccount(ctx, address)
		if rawAccount.GetPubKey() != nil || rawAccount.GetSequence() != 0 {
			return nil, fmt.Errorf("attempting to register an existing user account with address: %s", address.String())
		}

		switch account := rawAccount.(type) {
		case *authtypes.BaseAccount:
			rawAccount = &types.ForwardingAccount{
				BaseAccount: account,
				Channel:     msg.Channel,
				Recipient:   msg.Recipient,
				CreatedAt:   ctx.BlockHeight(),
			}
			k.authKeeper.SetAccount(ctx, rawAccount)

			k.IncrementNumOfAccounts(ctx, msg.Channel)
		case *types.ForwardingAccount:
			return nil, errors.New("account has already been registered")
		default:
			return nil, fmt.Errorf("unsupported account type: %T", rawAccount)
		}

		if !k.bankKeeper.GetAllBalances(ctx, address).IsZero() {
			account, ok := rawAccount.(*types.ForwardingAccount)
			if ok {
				k.SetPendingForward(ctx, account)
			}
		}

		return &types.MsgRegisterAccountResponse{Address: address.String()}, nil
	}

	account := types.ForwardingAccount{
		BaseAccount: authtypes.NewBaseAccountWithAddress(address),
		Channel:     msg.Channel,
		Recipient:   msg.Recipient,
		CreatedAt:   ctx.BlockHeight(),
	}

	k.authKeeper.SetAccount(ctx, &account)
	k.IncrementNumOfAccounts(ctx, msg.Channel)

	return &types.MsgRegisterAccountResponse{Address: address.String()}, nil
}

func (k *Keeper) ClearAccount(goCtx context.Context, msg *types.MsgClearAccount) (*types.MsgClearAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := sdk.MustAccAddressFromBech32(msg.Address)

	rawAccount := k.authKeeper.GetAccount(ctx, address)
	if rawAccount == nil {
		return nil, errors.New("account does not exist")
	}
	account, ok := rawAccount.(*types.ForwardingAccount)
	if !ok {
		return nil, errors.New("account is not a forwarding account")
	}

	if k.bankKeeper.GetAllBalances(ctx, address).IsZero() {
		return nil, errors.New("account does not require clearing")
	}

	k.SetPendingForward(ctx, account)

	return &types.MsgClearAccountResponse{}, nil
}
