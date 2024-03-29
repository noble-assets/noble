package app

import (
	"github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (app *App) DeliverTx(req abci.RequestDeliverTx) abci.ResponseDeliverTx {
	res := app.BaseApp.DeliverTx(req)
	ctx := app.BaseApp.DeliverState.Context()

	for _, event := range res.Events {
		err := app.FiatTokenFactoryKeeper.HandleDeliverTxEvent(ctx, event)
		if err != nil {
			app.SetDeliverState(ctx.BlockHeader())
			return errors.ResponseDeliverTxWithEvents(err, uint64(res.GasWanted), uint64(res.GasUsed), res.Events, app.Trace())
		}
	}

	return res
}
