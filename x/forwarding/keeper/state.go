package keeper

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/noble/v4/x/forwarding/types"
)

// PERSISTENT STATE

func (k *Keeper) GetNumOfAccounts(ctx sdk.Context, channel string) uint64 {
	key := types.NumOfAccountsKey(channel)
	bz := ctx.KVStore(k.storeKey).Get(key)

	if bz == nil {
		return 0
	} else {
		count, _ := strconv.ParseUint(string(bz), 10, 64)
		return count
	}
}

func (k *Keeper) GetAllNumOfAccounts(ctx sdk.Context) map[string]uint64 {
	counts := make(map[string]uint64)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.NumOfAccountsPrefix)
	iterator := sdk.KVStorePrefixIterator(store, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		count, _ := strconv.ParseUint(string(iterator.Value()), 10, 64)

		counts[channel] = count
	}

	return counts
}

func (k *Keeper) IncrementNumOfAccounts(ctx sdk.Context, channel string) {
	count := k.GetNumOfAccounts(ctx, channel)

	key := types.NumOfAccountsKey(channel)
	bz := []byte(strconv.Itoa(int(count + 1)))

	ctx.KVStore(k.storeKey).Set(key, bz)

	k.Logger(ctx).Info("registered a new account", "channel", channel)
}

func (k *Keeper) SetNumOfAccounts(ctx sdk.Context, channel string, count uint64) {
	key := types.NumOfAccountsKey(channel)
	bz := []byte(strconv.Itoa(int(count)))

	ctx.KVStore(k.storeKey).Set(key, bz)
}

func (k *Keeper) GetNumOfForwards(ctx sdk.Context, channel string) uint64 {
	key := types.NumOfForwardsKey(channel)
	bz := ctx.KVStore(k.storeKey).Get(key)

	if bz == nil {
		return 0
	} else {
		count, _ := strconv.ParseUint(string(bz), 10, 64)
		return count
	}
}

func (k *Keeper) GetAllNumOfForwards(ctx sdk.Context) map[string]uint64 {
	counts := make(map[string]uint64)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.NumOfForwardsPrefix)
	iterator := sdk.KVStorePrefixIterator(store, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		count, _ := strconv.ParseUint(string(iterator.Value()), 10, 64)

		counts[channel] = count
	}

	return counts
}

func (k *Keeper) IncrementNumOfForwards(ctx sdk.Context, channel string) {
	count := k.GetNumOfForwards(ctx, channel)

	key := types.NumOfForwardsKey(channel)
	bz := []byte(strconv.Itoa(int(count + 1)))

	ctx.KVStore(k.storeKey).Set(key, bz)
}

func (k *Keeper) SetNumOfForwards(ctx sdk.Context, channel string, count uint64) {
	key := types.NumOfForwardsKey(channel)
	bz := []byte(strconv.Itoa(int(count)))

	ctx.KVStore(k.storeKey).Set(key, bz)
}

func (k *Keeper) GetTotalForwarded(ctx sdk.Context, channel string) sdk.Coins {
	key := types.TotalForwardedKey(channel)
	bz := ctx.KVStore(k.storeKey).Get(key)

	total, _ := sdk.ParseCoinsNormalized(string(bz))
	return total
}

func (k *Keeper) GetAllTotalForwarded(ctx sdk.Context) map[string]string {
	totals := make(map[string]string)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TotalForwardedPrefix)
	iterator := sdk.KVStorePrefixIterator(store, nil)

	for ; iterator.Valid(); iterator.Next() {
		channel := string(iterator.Key())
		total := string(iterator.Value())

		totals[channel] = total
	}

	return totals
}

func (k *Keeper) IncrementTotalForwarded(ctx sdk.Context, channel string, coins sdk.Coins) {
	total := k.GetTotalForwarded(ctx, channel)

	key := types.TotalForwardedKey(channel)
	bz := []byte(total.Add(coins...).String())

	ctx.KVStore(k.storeKey).Set(key, bz)
}

func (k *Keeper) SetTotalForwarded(ctx sdk.Context, channel string, total sdk.Coins) {
	key := types.TotalForwardedKey(channel)
	bz := []byte(total.String())

	ctx.KVStore(k.storeKey).Set(key, bz)
}

// TRANSIENT STATE

func (k *Keeper) GetPendingForwards(ctx sdk.Context) (accounts []types.ForwardingAccount) {
	itr := ctx.TransientStore(k.transientKey).Iterator(types.PendingForwardsPrefix, nil)

	for ; itr.Valid(); itr.Next() {
		var account types.ForwardingAccount
		k.cdc.MustUnmarshal(itr.Value(), &account)

		accounts = append(accounts, account)
	}

	return
}

func (k *Keeper) HasPendingForward(ctx sdk.Context, account *types.ForwardingAccount) bool {
	key := types.PendingForwardsKey(account)
	bz := ctx.TransientStore(k.transientKey).Get(key)

	return bz != nil
}

func (k *Keeper) SetPendingForward(ctx sdk.Context, account *types.ForwardingAccount) {
	if k.HasPendingForward(ctx, account) {
		return
	}

	key := types.PendingForwardsKey(account)
	bz := k.cdc.MustMarshal(account)

	ctx.TransientStore(k.transientKey).Set(key, bz)
}
