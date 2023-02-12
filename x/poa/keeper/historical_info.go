package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	value := store.Get(key)
	if value == nil {
		return stakingtypes.HistoricalInfo{}, false
	}

	return stakingtypes.MustUnmarshalHistoricalInfo(types.ModuleCdc, value), true
}

// SetHistoricalInfo sets the historical info at a given height
func (k Keeper) SetHistoricalInfo(ctx sdk.Context, height int64, hi *stakingtypes.HistoricalInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)
	value := types.ModuleCdc.MustMarshal(hi)
	store.Set(key, value)
}

// DeleteHistoricalInfo deletes the historical info at a given height
func (k Keeper) DeleteHistoricalInfo(ctx sdk.Context, height int64) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	store.Delete(key)
}

// IterateHistoricalInfo provides an interator over all stored HistoricalInfo
//
//	objects. For each HistoricalInfo object, cb will be called. If the cb returns
//
// true, the iterator will close and stop.
func (k Keeper) IterateHistoricalInfo(ctx sdk.Context, cb func(stakingtypes.HistoricalInfo) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.HistoricalInfoKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		histInfo := stakingtypes.MustUnmarshalHistoricalInfo(types.ModuleCdc, iterator.Value())
		if cb(histInfo) {
			break
		}
	}
}

// GetAllHistoricalInfo returns all stored HistoricalInfo objects.
func (k Keeper) GetAllHistoricalInfo(ctx sdk.Context) []stakingtypes.HistoricalInfo {
	var infos []stakingtypes.HistoricalInfo

	k.IterateHistoricalInfo(ctx, func(histInfo stakingtypes.HistoricalInfo) bool {
		infos = append(infos, histInfo)
		return false
	})

	return infos
}

// TrackHistoricalInfo saves the latest historical-info and deletes the oldest
// heights that are below pruning height
func (k Keeper) TrackHistoricalInfo(ctx sdk.Context) {
	entryNum := k.HistoricalEntries(ctx)

	// Prune store to ensure we only have parameter-defined historical entries.
	// In most cases, this will involve removing a single historical entry.
	// In the rare scenario when the historical entries gets reduced to a lower value k'
	// from the original value k. k - k' entries must be deleted from the store.
	// Since the entries to be deleted are always in a continuous range, we can iterate
	// over the historical entries starting from the most recent version to be pruned
	// and then return at the first empty entry.
	for i := ctx.BlockHeight() - int64(entryNum); i >= 0; i-- {
		_, found := k.GetHistoricalInfo(ctx, i)
		if found {
			k.DeleteHistoricalInfo(ctx, i)
		} else {
			break
		}
	}

	// if there is no need to persist historicalInfo, return
	if entryNum == 0 {
		return
	}

	// Create HistoricalInfo struct
	lastVals := k.GetAllValidatorsStaking(ctx)
	historicalEntry := stakingtypes.NewHistoricalInfo(ctx.BlockHeader(), lastVals, k.PowerReduction(ctx))

	// Set latest HistoricalInfo at current height
	k.SetHistoricalInfo(ctx, ctx.BlockHeight(), &historicalEntry)
}
