package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/v4/x/tokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/x/tokenfactory/keeper"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))
)

func SimulateMsgUpdateOwner(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k *keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUpdateOwner{
			From: simAccount.Address.String(),
		}

		// TODO: Handling the UpdateOwner simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UpdateOwner simulation not implemented"), nil, nil
	}
}
