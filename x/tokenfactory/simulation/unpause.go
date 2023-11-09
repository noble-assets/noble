package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/types"
)

func SimulateMsgUnpause(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUnpause{
			From: simAccount.Address.String(),
		}

		// TODO: Handling the Unpause simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "Unpause simulation not implemented"), nil, nil
	}
}
