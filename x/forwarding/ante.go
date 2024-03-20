package forwarding

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/noble-assets/noble/v5/x/forwarding/keeper"
	"github.com/noble-assets/noble/v5/x/forwarding/types"
)

type Decorator struct {
	authKeeper ante.AccountKeeper
	keeper     *keeper.Keeper
}

var _ sdk.AnteDecorator = Decorator{}

func NewAnteDecorator(keeper *keeper.Keeper, authKeeper ante.AccountKeeper) Decorator {
	return Decorator{
		authKeeper: authKeeper,
		keeper:     keeper,
	}
}

func (d Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()

	err = d.CheckMessages(ctx, msgs)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (d Decorator) CheckMessages(ctx sdk.Context, msgs []sdk.Msg) error {
	for _, raw := range msgs {
		if msg, ok := raw.(*authz.MsgExec); ok {
			nestedMsgs, err := msg.GetMessages()
			if err != nil {
				return err
			}

			return d.CheckMessages(ctx, nestedMsgs)
		}

		switch msg := raw.(type) {
		// TODO: Add CCTP message check back!
		case *banktypes.MsgMultiSend:
			for _, output := range msg.Outputs {
				address := sdk.MustAccAddressFromBech32(output.Address)

				rawAccount := d.authKeeper.GetAccount(ctx, address)
				if rawAccount == nil {
					continue
				}

				account, ok := rawAccount.(*types.ForwardingAccount)
				if !ok {
					continue
				}

				d.keeper.SetPendingForward(ctx, account)
			}
		case *banktypes.MsgSend:
			address := sdk.MustAccAddressFromBech32(msg.ToAddress)

			rawAccount := d.authKeeper.GetAccount(ctx, address)
			if rawAccount == nil {
				return nil
			}

			account, ok := rawAccount.(*types.ForwardingAccount)
			if !ok {
				return nil
			}

			d.keeper.SetPendingForward(ctx, account)
			// TODO: Add FiatTokenFactory message check back!
		}
	}

	return nil
}
