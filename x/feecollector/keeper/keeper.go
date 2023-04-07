package keeper

import (
	"github.com/strangelove-ventures/noble/x/feecollector/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		paramstore       paramtypes.Subspace
		authKeeper       types.AccountKeeper
		bankKeeper       types.BankKeeper
		feeCollectorName string // name of the FeeCollector ModuleAccount
	}
)

func NewKeeper(
	ps paramtypes.Subspace,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	feeCollectorName string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		paramstore:       ps,
		authKeeper:       authKeeper,
		bankKeeper:       bankKeeper,
		feeCollectorName: feeCollectorName,
	}
}
