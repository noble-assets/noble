package keeper

import (
	"encoding/binary"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetTokenMessenger sets an token messenger in the store
func (k Keeper) SetTokenMessenger(ctx sdk.Context, tokenMessenger types.TokenMessenger) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&tokenMessenger)
	store.Set(types.KeyPrefix(types.TokenMessengerKeyPrefix), b)
}

// GetTokenMessenger returns tokenMessenger
func (k Keeper) GetTokenMessenger(ctx sdk.Context, domainId uint32) (val types.TokenMessenger, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenMessengerKeyPrefix))

	b := store.Get(types.KeyPrefix(string(types.TokenMessengerKey([]byte(strconv.Itoa(int(domainId)))))))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// DeleteTokenMessenger removes a token messenger from the store
func (k Keeper) DeleteTokenMessenger(
	ctx sdk.Context,
	domainId uint32,
) {
	var key []byte
	binary.LittleEndian.PutUint32(key, domainId)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenMessengerKeyPrefix))
	store.Delete(types.TokenMessengerKey(
		key,
	))
}

// GetAllTokenMessengers returns all token messengers
func (k Keeper) GetAllTokenMessengers(ctx sdk.Context) (list []types.TokenMessenger) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TokenMessengerKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.TokenMessenger
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
