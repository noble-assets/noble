package simulation

import (
	"math/rand"

	"github.com/strangelove-ventures/noble/x/fiattokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/fiattokenfactory/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgUpdateMasterMinter(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUpdateMasterMinter{
			From: simAccount.Address.String(),
		}

		// TODO: Handling the UpdateMasterMinter simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UpdateMasterMinter simulation not implemented"), nil, nil
	}
}
