package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	"github.com/noble-assets/noble/v4/x/forwarding/types"
	"github.com/tendermint/tendermint/libs/log"
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
func (k *Keeper) ExecuteForwards(ctx sdk.Context) {
	forwards := k.GetPendingForwards(ctx)
	if len(forwards) > 0 {
		k.Logger(ctx).Info(fmt.Sprintf("executing %d automatic forward(s)", len(forwards)))
	}

	for _, forward := range forwards {
		balances := k.bankKeeper.GetAllBalances(ctx, forward.GetAddress())

		for _, balance := range balances {
			timeout := uint64(ctx.BlockTime().UnixNano()) + transfertypes.DefaultRelativePacketTimeoutTimestamp
			err := k.transferKeeper.SendTransfer(ctx, transfertypes.PortID, forward.Channel, balance, forward.GetAddress(), forward.Recipient, clienttypes.ZeroHeight(), timeout)
			if err != nil {
				// TODO: Consider moving to persistent store in order to retry in future blocks?
				k.Logger(ctx).Error("unable to execute automatic forward", "channel", forward.Channel, "address", forward.GetAddress().String(), "amount", balance.String(), "err", err)
			} else {
				k.IncrementNumOfForwards(ctx, forward.Channel)
				k.IncrementTotalForwarded(ctx, forward.Channel, balance)
			}
		}
	}

	// NOTE: As pending forwards are stored in transient state, they are automatically cleared at the end of the block lifecycle. No further action is required.
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", types.ModuleName)
}
