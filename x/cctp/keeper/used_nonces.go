package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/strangelove-ventures/noble/x/cctp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type UsedNonce struct {
	Nonce        uint64
	SourceDomain uint32
}

func (n *UsedNonce) Key() []byte {
	sourceDomainBz := make([]byte, 4)
	binary.BigEndian.PutUint32(sourceDomainBz, n.SourceDomain)

	nonceBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBz, n.Nonce)

	return append(sourceDomainBz, nonceBz...)
}

func (n *UsedNonce) Unmarshal(bz []byte) error {
	if len(bz) != 12 {
		return fmt.Errorf("used nonce length %d is invalid, must be 12", len(bz))
	}

	n.SourceDomain = binary.BigEndian.Uint32(bz[0:4])
	n.Nonce = binary.BigEndian.Uint64(bz[4:12])

	return nil
}

// SetUsedNonce sets a nonce in the store
func (k Keeper) SetUsedNonce(ctx sdk.Context, key UsedNonce) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))
	store.Set(key.Key(), []byte{1})
}

// GetUsedNonce returns nonce
func (k Keeper) GetUsedNonce(ctx sdk.Context, key UsedNonce) (found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))

	b := store.Get(key.Key())
	if b == nil {
		return false
	}

	return true
}

// GetAllUsedNonces returns all UsedNonces
func (k Keeper) GetAllUsedNonces(ctx sdk.Context) (list []UsedNonce) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UsedNonceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val UsedNonce
		key := iterator.Key()
		if err := val.Unmarshal(key); err != nil {
			panic(err)
		}

		list = append(list, val)
	}

	return
}
