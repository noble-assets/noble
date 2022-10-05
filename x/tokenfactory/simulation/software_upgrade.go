package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"noble/x/tokenfactory/keeper"
	"noble/x/tokenfactory/types"
)

func SimulateMsgSoftwareUpgrade(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgSoftwareUpgrade{
			From: simAccount.Address.String(),
		}

		// TODO: Handling the SoftwareUpgrade simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "SoftwareUpgrade simulation not implemented"), nil, nil
	}
}
