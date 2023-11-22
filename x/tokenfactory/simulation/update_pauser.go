package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/noble-assets/noble/v5/x/tokenfactory/keeper"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
)

func SimulateMsgUpdatePauser(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUpdatePauser{
			From: simAccount.Address.String(),
		}

		// TODO: Handling the UpdatePauser simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UpdatePauser simulation not implemented"), nil, nil
	}
}
