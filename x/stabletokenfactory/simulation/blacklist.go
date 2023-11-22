package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/noble-assets/noble/v4/x/stabletokenfactory/keeper"
	"github.com/noble-assets/noble/v4/x/stabletokenfactory/types"
)

func SimulateMsgBlacklist(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgBlacklist{
			From: simAccount.Address.String(),
		}

		// TODO: Handling the Blacklist simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "Blacklist simulation not implemented"), nil, nil
	}
}
