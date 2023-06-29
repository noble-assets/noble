package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AllocateTokens(ctx sdk.Context) {
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	foundAmountGreaterThanZero := false
	for _, coin := range feesCollectedInt {
		if coin.Amount.GT(sdk.ZeroInt()) {
			foundAmountGreaterThanZero = true
			break
		}
	}
	if !foundAmountGreaterThanZero {
		// no fees to distribute
		return
	}
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	params := k.GetParams(ctx)
	feesToDistribute := feesCollected.MulDecTruncate(params.Share)

	foundAmountGreaterThanZero = false
	for _, coin := range feesToDistribute {
		truncated, _ := coin.TruncateDecimal()
		if truncated.Amount.GT(sdk.ZeroInt()) {
			foundAmountGreaterThanZero = true
			break
		}
	}
	if !foundAmountGreaterThanZero {
		// no fees to distribute
		return
	}

	for _, d := range params.DistributionEntities {
		entityShare := feesToDistribute.MulDecTruncate(d.Share)

		var coins sdk.Coins

		for _, s := range entityShare {
			truncated, _ := s.TruncateDecimal()
			if truncated.Amount.GT(sdk.ZeroInt()) {
				coins = append(coins, truncated)
			}
		}

		if len(coins) == 0 {
			continue
		}

		acc := sdk.MustAccAddressFromBech32(d.Address)

		// transfer collected fees to the distribution entity account
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, acc, coins)
		if err != nil {
			ctx.Logger().Error("Error allocating tokens to distribution entity: %s "+err.Error(), d.Address)
		}
	}
}
