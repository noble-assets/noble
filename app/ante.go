package app

import (
	"errors"

	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
)

type HandlerOptions struct {
	ante.HandlerOptions

	cdc                    codec.Codec
	FiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper
	IBCKeeper              *ibckeeper.Keeper
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.New("account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errors.New("bank keeper is required for ante builder")
	}

	if options.FiatTokenFactoryKeeper == nil {
		return nil, errors.New("fiat token factory keeper is required for ante builder")
	}

	if options.IBCKeeper == nil {
		return nil, errors.New("ibc keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errors.New("sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		fiattokenfactory.NewIsBlacklistedDecorator(options.FiatTokenFactoryKeeper),
		fiattokenfactory.NewIsPausedDecorator(options.cdc, options.FiatTokenFactoryKeeper),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
