package keeper

import (
	"fmt"

<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMintingDenom set mintingDenom in the store
func (k *Keeper) SetMintingDenom(ctx sdk.Context, mintingDenom types.MintingDenom) {
	if k.MintingDenomSet(ctx) {
		panic(types.ErrMintingDenomSet)
	}

	_, found := k.bankKeeper.GetDenomMetaData(ctx, mintingDenom.Denom)
	if !found {
		panic(fmt.Sprintf("Denom metadata for '%s' should be set", mintingDenom.Denom))
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))
	b := k.cdc.MustMarshal(&mintingDenom)
	store.Set(types.KeyPrefix(types.MintingDenomKey), b)
}

// GetMintingDenom returns mintingDenom
func (k *Keeper) GetMintingDenom(ctx sdk.Context) (val types.MintingDenom) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))

	b := store.Get(types.KeyPrefix(types.MintingDenomKey))
	if b == nil {
		panic("Minting denom is not set")
	}

	k.cdc.MustUnmarshal(b, &val)
	return val
}

// MintingDenomSet returns true if the MintingDenom is already set in the store, it returns false otherwise.
func (k Keeper) MintingDenomSet(ctx sdk.Context) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))

	b := store.Get(types.KeyPrefix(types.MintingDenomKey))

	return b != nil
}
