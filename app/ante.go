package app

import (
	tokenfactory "noble/x/tokenfactory/keeper"
	tokenfactorytypes "noble/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibcante "github.com/cosmos/ibc-go/v5/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v5/modules/core/keeper"
)

type HandlerOptions struct {
	ante.HandlerOptions
	tokenfactorykeeper tokenfactory.Keeper
	IBCKeeper          *ibckeeper.Keeper
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
		switch m.(type) {
		case *banktypes.MsgSend, *banktypes.MsgMultiSend, *transfertypes.MsgTransfer:
			paused := ad.tokenfactory.GetPaused(ctx)
			if paused.Paused {
				return ctx, sdkerrors.Wrapf(tokenfactorytypes.ErrPaused, "can not perform token transfers")
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

type IsBlacklistedDecorator struct {
	tokenfactory tokenfactory.Keeper
}

func NewIsBlacklistedDecorator(tk tokenfactory.Keeper) IsBlacklistedDecorator {
	return IsBlacklistedDecorator{
		tokenfactory: tk,
	}
}

func (ad IsBlacklistedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch m := m.(type) {
		case *banktypes.MsgSend, *banktypes.MsgMultiSend, *transfertypes.MsgTransfer:
			var addresses []string
			switch m := m.(type) {
			case *banktypes.MsgSend:
				addresses = append(addresses, m.FromAddress)
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					addresses = append(addresses, i.Address)
				}
			case *transfertypes.MsgTransfer:
				addresses = append(addresses, m.Sender)
			}
			for _, address := range addresses {
				_, found := ad.tokenfactory.GetBlacklisted(ctx, address)
				if found {
					return ctx, sdkerrors.Wrapf(tokenfactorytypes.ErrUnauthorized, "an address (%s) is blacklisted and can not send or receive tokens", address)
				}
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
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for AnteHandler")
	}
	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for AnteHandler")
	}
	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}
	sigGasConsumer := options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		NewIsPausedDecorator(options.tokenfactorykeeper),
		NewIsBlacklistedDecorator(options.tokenfactorykeeper),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}
	return sdk.ChainAnteDecorators(anteDecorators...), nil

}
