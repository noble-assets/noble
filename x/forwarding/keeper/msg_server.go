package keeper

import (
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/noble-assets/noble/v4/x/forwarding/types"
)

var _ types.MsgServer = &Keeper{}

func (k *Keeper) RegisterAccount(goCtx context.Context, msg *types.MsgRegisterAccount) (*types.MsgRegisterAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address := types.GenerateAddress(msg.Channel, msg.Recipient)

	if _, found := k.channelKeeper.GetChannel(ctx, transfertypes.PortID, msg.Channel); !found {
		return nil, fmt.Errorf("channel does not exist: %s", msg.Channel)
	}

	if k.authKeeper.HasAccount(ctx, address) {
		switch account := k.authKeeper.GetAccount(ctx, address).(type) {
		case *authtypes.BaseAccount:
			k.authKeeper.SetAccount(ctx, &types.ForwardingAccount{
				BaseAccount: account,
				Channel:     msg.Channel,
				Recipient:   msg.Recipient,
				CreatedAt:   ctx.BlockHeight(),
			})

			k.IncrementNumOfAccounts(ctx, msg.Channel)
		case *types.ForwardingAccount:
			return nil, errors.New("account has already been registered")
		default:
			break
		}

		if !k.bankKeeper.GetAllBalances(ctx, address).IsZero() {
			rawAccount := k.authKeeper.GetAccount(ctx, address)
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
