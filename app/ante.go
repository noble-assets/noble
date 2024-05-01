package app

import (
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	fiattokenfactorykeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
<<<<<<< HEAD
	"github.com/noble-assets/noble/v4/x/forwarding"
	forwardingkeeper "github.com/noble-assets/noble/v4/x/forwarding/keeper"
	feeante "github.com/noble-assets/noble/v4/x/globalfee/ante"
=======
	"github.com/noble-assets/forwarding/x/forwarding"
	forwardingkeeper "github.com/noble-assets/forwarding/x/forwarding/keeper"
	feeante "github.com/noble-assets/noble/v5/x/globalfee/ante"
>>>>>>> ee651ba (refactor: use migrated `x/forwarding` (#357))
)

type HandlerOptions struct {
	ante.HandlerOptions
	cdc                    codec.Codec
	fiatTokenFactoryKeeper *fiattokenfactorykeeper.Keeper
	IBCKeeper              *ibckeeper.Keeper
	GlobalFeeSubspace      paramtypes.Subspace
	StakingSubspace        paramtypes.Subspace
	ForwardingKeeper       *forwardingkeeper.Keeper
}

// maxTotalBypassMinFeeMsgGasUsage is the allowed maximum gas usage
// for all the bypass msgs in a transactions.
// A transaction that contains only bypass message types and the gas usage does not
// exceed maxTotalBypassMinFeeMsgGasUsage can be accepted with a zero fee.
// For details, see gaiafeeante.NewFeeDecorator()
var maxTotalBypassMinFeeMsgGasUsage uint64 = 1_000_000

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
		ante.NewRejectExtensionOptionsDecorator(),
		fiattokenfactory.NewIsBlacklistedDecorator(options.fiatTokenFactoryKeeper),
		fiattokenfactory.NewIsPausedDecorator(options.cdc, options.fiatTokenFactoryKeeper),
		forwarding.NewAnteDecorator(options.ForwardingKeeper, options.AccountKeeper),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		feeante.NewFeeDecorator(options.GlobalFeeSubspace, options.StakingSubspace, maxTotalBypassMinFeeMsgGasUsage),

		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewAnteDecorator(options.IBCKeeper),
	}
	return sdk.ChainAnteDecorators(anteDecorators...), nil

}
