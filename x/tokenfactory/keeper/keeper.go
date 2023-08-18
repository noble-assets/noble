package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
	}
}

// ValidatePrivileges checks if a specified address has already been assigned to a privileged role.
func (k Keeper) ValidatePrivileges(ctx sdk.Context, address string) error {
	acc, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return err
	}

	owner, found := k.GetOwner(ctx)
	if found && owner.Address == acc.String() {
		return sdkerrors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to owner role", acc.String())
	}

	blacklister, found := k.GetBlacklister(ctx)
	if found && blacklister.Address == acc.String() {
		return sdkerrors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to black lister role", acc.String())
	}

	masterminter, found := k.GetMasterMinter(ctx)
	if found && masterminter.Address == acc.String() {
		return sdkerrors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to master minter role", acc.String())
	}

	pauser, found := k.GetPauser(ctx)
	if found && pauser.Address == acc.String() {
		return sdkerrors.Wrapf(types.ErrAlreadyPrivileged, "cannot assign (%s) to pauser role", acc.String())
	}

	return nil
}
