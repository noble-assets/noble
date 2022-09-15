package app

import (
	tokenfactory "noble/x/tokenfactory/keeper"
	tokenfactorytypes "noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
)

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
func NewAnteHandler(
	ak authkeeper.AccountKeeper,
	bankKeeper types.BankKeeper,
	feegrantKeeper ante.FeegrantKeeper,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler signing.SignModeHandler,
	tokenfactoryKeeper tokenfactory.Keeper,
) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewIsPausedDecorator(tokenfactoryKeeper),
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewValidateBasicDecorator(),
		ante.TxTimeoutHeightDecorator{},
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		ante.NewSigVerificationDecorator(ak, signModeHandler),
		ante.NewIncrementSequenceDecorator(ak),
	)
}
