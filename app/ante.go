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
)

type IsPausedDecorator struct {
	ak           authkeeper.AccountKeeper
	bankKeeper   types.BankKeeper
	tokenfactory tokenfactory.Keeper
}

func NewIsPausedDecorator(ak authkeeper.AccountKeeper, bk types.BankKeeper, tk tokenfactory.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		ak:           ak,
		bankKeeper:   bk,
		tokenfactory: tk,
	}
}

func (ad IsPausedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		_, ok := m.(*banktypes.MsgSend)
		if !ok {
			continue
		}
		paused, _ := ad.tokenfactory.GetPaused(ctx)
		if paused.Paused {
			return ctx, sdkerrors.Wrapf(tokenfactorytypes.ErrBurn, "minter address is blacklisted")
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
		NewIsPausedDecorator(ak, bankKeeper, tokenfactoryKeeper),
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
