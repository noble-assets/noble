package keeper

import (
	"fmt"
	"noble/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SetMintingDenom set mintingDenom in the store
func (k Keeper) SetMintingDenom(ctx sdk.Context, mintingDenom types.MintingDenom) {
	denom := mintingDenom.Denom
	metadata := banktypes.Metadata{
		Base: fmt.Sprintf("u%s", denom),
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    fmt.Sprintf("u%s", denom),
				Exponent: 0,
			},
			{
				Denom:    fmt.Sprintf("m%s", denom),
				Exponent: 3,
			},
			{
				Denom:    denom,
				Exponent: 6,
			},
		},
	}
	mintingDenom.Denom = fmt.Sprintf("u%s", denom)
	k.bankKeeper.SetDenomMetaData(ctx, metadata)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))
	b := k.cdc.MustMarshal(&mintingDenom)
	store.Set(types.KeyPrefix(types.MintingDenomKey), b)
}

// GetMintingDenom returns mintingDenom
func (k Keeper) GetMintingDenom(ctx sdk.Context) (val types.MintingDenom) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))

	b := store.Get(types.KeyPrefix(types.MintingDenomKey))
	if b == nil {
		panic("Minting denom is not set")
	}

	k.cdc.MustUnmarshal(b, &val)
	return val
}

// RemoveMintingDenom removes mintingDenom from the store
func (k Keeper) RemoveMintingDenom(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MintingDenomKey))
	store.Delete(types.KeyPrefix(types.MintingDenomKey))
}
