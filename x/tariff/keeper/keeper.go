package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"

	// IBC Core
	portTypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	// Tariff
	"github.com/strangelove-ventures/noble/x/tariff/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storeTypes.StoreKey

		authority        string
		feeCollectorName string // name of the FeeCollector ModuleAccount

		authKeeper  types.AccountKeeper
		bankKeeper  types.BankKeeper
		ics4Wrapper portTypes.ICS4Wrapper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storeTypes.StoreKey,
	authority string,
	feeCollectorName string,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	ics4Wrapper portTypes.ICS4Wrapper,
) *Keeper {
	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		authority:        authority,
		feeCollectorName: feeCollectorName,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		ics4Wrapper:      ics4Wrapper,
	}
}

func (k *Keeper) SetICS4Wrapper(ics4Wrapper portTypes.ICS4Wrapper) {
	k.ics4Wrapper = ics4Wrapper
}
