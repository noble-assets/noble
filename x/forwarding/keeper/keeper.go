package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/noble-assets/noble/v5/x/forwarding/types"
)

type Keeper struct {
	cdc          codec.Codec
	storeKey     storetypes.StoreKey
	transientKey *storetypes.TransientStoreKey

	authKeeper     types.AccountKeeper
	bankKeeper     types.BankKeeper
	channelKeeper  types.ChannelKeeper
	transferKeeper types.TransferKeeper
}

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	transientKey *storetypes.TransientStoreKey,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	transferKeeper types.TransferKeeper,
) *Keeper {
	return &Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		transientKey: transientKey,

		authKeeper:     authKeeper,
		bankKeeper:     bankKeeper,
		channelKeeper:  channelKeeper,
		transferKeeper: transferKeeper,
	}
}

// ExecuteForwards is an end block hook that clears all pending forwards from transient state.
func (k *Keeper) ExecuteForwards(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	forwards := k.GetPendingForwards(sdkCtx)
	if len(forwards) > 0 {
		k.Logger(sdkCtx).Info(fmt.Sprintf("executing %d automatic forward(s)", len(forwards)))
	}

	for _, forward := range forwards {
		channel, _ := k.channelKeeper.GetChannel(sdkCtx, transfertypes.PortID, forward.Channel)
		if channel.State != channeltypes.OPEN {
			k.Logger(sdkCtx).Error("skipped automatic forward due to non open channel", "channel", forward.Channel, "address", forward.GetAddress().String(), "state", channel.State.String())
			continue
		}

		balances := k.bankKeeper.GetAllBalances(sdkCtx, forward.GetAddress())

		for _, balance := range balances {
			timeout := uint64(sdkCtx.BlockTime().UnixNano()) + transfertypes.DefaultRelativePacketTimeoutTimestamp

			_, err := k.transferKeeper.Transfer(ctx, &transfertypes.MsgTransfer{
				SourcePort:       transfertypes.PortID,
				SourceChannel:    forward.Channel,
				Token:            balance,
				Sender:           forward.GetAddress().String(),
				Receiver:         forward.Recipient,
				TimeoutHeight:    clienttypes.ZeroHeight(),
				TimeoutTimestamp: timeout,
				Memo:             "",
			})
			if err != nil {
				// TODO: Consider moving to persistent store in order to retry in future blocks?
				k.Logger(sdkCtx).Error("unable to execute automatic forward", "channel", forward.Channel, "address", forward.GetAddress().String(), "amount", balance.String(), "err", err)
			} else {
				k.IncrementNumOfForwards(sdkCtx, forward.Channel)
				k.IncrementTotalForwarded(sdkCtx, forward.Channel, balance)
			}
		}
	}

	// NOTE: As pending forwards are stored in transient state, they are automatically cleared at the end of the block lifecycle. No further action is required.
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}
