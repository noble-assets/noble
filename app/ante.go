package app

import (
	tokenfactory "noble/x/tokenfactory/keeper"
	tokenfactorytypes "noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
)

type HandlerOptions struct {
	ante.HandlerOptions
	tokenfactorykeeper tokenfactory.Keeper
}

type IsPausedDecorator struct {
	tokenfactory tokenfactory.Keeper
}

func NewIsPausedDecorator(tk tokenfactory.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		tokenfactory: tk,
	}
}

func (ad IsPausedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch m := m.(type) {
		case *banktypes.MsgSend, *transfertypes.MsgTransfer:
			paused, _ := ad.tokenfactory.GetPaused(ctx)
			if paused.Paused {
				return ctx, sdkerrors.Wrapf(tokenfactorytypes.ErrPaused, "can not perform token transfers")
			}
			var address string
			switch m := m.(type) {
			case *banktypes.MsgSend:
				address = m.FromAddress
			case *transfertypes.MsgTransfer:
				address = m.Sender
			}
			_, found := ad.tokenfactory.GetBlacklisted(ctx, address)
			if found {
				return ctx, sdkerrors.Wrapf(tokenfactorytypes.ErrUnauthorized, "an account is blacklisted and can not transfer tokens")
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer
func NewAnteHandler(opts HandlerOptions) (sdk.AnteHandler, error) {
	if opts.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if opts.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if opts.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	sigGasConsumer := opts.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		NewIsPausedDecorator(opts.tokenfactorykeeper),
		ante.NewValidateBasicDecorator(),
		ante.TxTimeoutHeightDecorator{},
		ante.NewValidateMemoDecorator(opts.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(opts.AccountKeeper),
		ante.NewSetPubKeyDecorator(opts.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(opts.AccountKeeper),
		ante.NewSigGasConsumeDecorator(opts.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(opts.AccountKeeper, opts.SignModeHandler),
		ante.NewIncrementSequenceDecorator(opts.AccountKeeper),
	}
	return sdk.ChainAnteDecorators(anteDecorators...), nil

}
