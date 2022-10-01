package keeper

import (
	"noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMintingDenom set mintingDenom in the store
func (k Keeper) SetMintingDenom(ctx sdk.Context, mintingDenom types.MintingDenom) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))
	b := k.cdc.MustMarshal(&mintingDenom)
	store.Set([]byte{0}, b)
}

// GetMintingDenom returns mintingDenom
func (k Keeper) GetMintingDenom(ctx sdk.Context) (val types.MintingDenom) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))

	b := store.Get([]byte{0})
	if b == nil {
		panic("Minting denom is not set")
	}

	k.cdc.MustUnmarshal(b, &val)
	return val
}

// RemoveMintingDenom removes mintingDenom from the store
func (k Keeper) RemoveMintingDenom(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))
	store.Delete([]byte{0})
}
