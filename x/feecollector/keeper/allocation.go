package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AllocateTokens(ctx sdk.Context) {
	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// calculate fraction allocated to validators
	params := k.GetParams(ctx)
	feesToDistribute := feesCollected.MulDecTruncate(params.Share)

	for _, d := range params.DistributionEntities {
		entityShare := feesToDistribute.MulDecTruncate(d.Share)
		// entityShareInt := sdk.NewDecCoinFromDec()
		x := sdk.NewCoin(entityShare)

		acc := sdk.MustAccAddressFromBech32(d.Address)

		// transfer collected fees to the distribution module account
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, k.feeCollectorName, acc, entityShare)
		if err != nil {
			panic(err)
		}

	}

}
