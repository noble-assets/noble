package app

import (
	fiattokenfactory "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	stabletokenfactorykeeper "github.com/strangelove-ventures/noble/v5/x/stabletokenfactory/keeper"
	stabletokenfactorytypes "github.com/strangelove-ventures/noble/v5/x/stabletokenfactory/types"
	tokenfactory "github.com/strangelove-ventures/noble/v5/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/strangelove-ventures/noble/v5/x/tokenfactory/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	transfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"

	feeante "github.com/strangelove-ventures/noble/v5/x/globalfee/ante"
)

type HandlerOptions struct {
	ante.HandlerOptions
	tokenFactoryKeeper       *tokenfactory.Keeper
	fiatTokenFactoryKeeper   *fiattokenfactory.Keeper
	stableTokenFactoryKeeper *stabletokenfactorykeeper.Keeper
	IBCKeeper                *ibckeeper.Keeper
	GlobalFeeSubspace        paramtypes.Subspace
	StakingSubspace          paramtypes.Subspace
}

type IsPausedDecorator struct {
	tokenFactory       *tokenfactory.Keeper
	fiatTokenFactory   *fiattokenfactory.Keeper
	stableTokenFactory *stabletokenfactorykeeper.Keeper
}

func NewIsPausedDecorator(tf *tokenfactory.Keeper, ctf *fiattokenfactory.Keeper, stf *stabletokenfactorykeeper.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		tokenFactory:       tf,
		fiatTokenFactory:   ctf,
		stableTokenFactory: stf,
	}
}

func (ad IsPausedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch m := m.(type) {
		case *banktypes.MsgSend, *banktypes.MsgMultiSend, *transfertypes.MsgTransfer:
			switch m := m.(type) {
			case *banktypes.MsgSend:
				for _, c := range m.Amount {
					paused, err := checkPausedStatebyTokenFactory(ctx, c, ad.tokenFactory, ad.fiatTokenFactory, ad.stableTokenFactory)
					if paused {
						return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						paused, err := checkPausedStatebyTokenFactory(ctx, c, ad.tokenFactory, ad.fiatTokenFactory, ad.stableTokenFactory)
						if paused {
							return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
						}
					}
				}
			case *transfertypes.MsgTransfer:
				paused, err := checkPausedStatebyTokenFactory(ctx, m.Token, ad.tokenFactory, ad.fiatTokenFactory, ad.stableTokenFactory)
				if paused {
					return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
				}
			default:
				continue
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

func checkPausedStatebyTokenFactory(ctx sdk.Context, c sdk.Coin, tf *tokenfactory.Keeper, ctf *fiattokenfactory.Keeper, stf *stabletokenfactorykeeper.Keeper) (bool, *sdkerrors.Error) {
	tfMintingDenom := tf.GetMintingDenom(ctx)
	if c.Denom == tfMintingDenom.Denom {
		paused := tf.GetPaused(ctx)
		if paused.Paused {
			return true, tokenfactorytypes.ErrPaused
		}
	}
	ctfMintingDenom := ctf.GetMintingDenom(ctx)
	if c.Denom == ctfMintingDenom.Denom {
		paused := ctf.GetPaused(ctx)
		if paused.Paused {
			return true, fiattokenfactorytypes.ErrPaused
		}
	}
	stfMintingDenom := stf.GetMintingDenom(ctx)
	if c.Denom == stfMintingDenom.Denom {
		paused := stf.GetPaused(ctx)
		if paused.Paused {
			return true, stabletokenfactorytypes.ErrPaused
		}
	}
	return false, nil
}

type IsBlacklistedDecorator struct {
	tokenfactory       *tokenfactory.Keeper
	fiattokenfactory   *fiattokenfactory.Keeper
	stabletokenfactory *stabletokenfactorykeeper.Keeper
}

func NewIsBlacklistedDecorator(tf *tokenfactory.Keeper, ctf *fiattokenfactory.Keeper, stf *stabletokenfactorykeeper.Keeper) IsBlacklistedDecorator {
	return IsBlacklistedDecorator{
		tokenfactory:       tf,
		fiattokenfactory:   ctf,
		stabletokenfactory: stf,
	}
}

func (ad IsBlacklistedDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch m := m.(type) {
		case *banktypes.MsgSend, *banktypes.MsgMultiSend, *transfertypes.MsgTransfer:
			switch m := m.(type) {
			case *banktypes.MsgSend:
				for _, c := range m.Amount {
					addresses := []string{m.ToAddress, m.FromAddress}
					blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, c, ad.tokenfactory, ad.fiattokenfactory, ad.stabletokenfactory)
					if blacklisted {
						return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
					}
					if err != nil {
						return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						addresses := []string{i.Address}
						blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, c, ad.tokenfactory, ad.fiattokenfactory, ad.stabletokenfactory)
						if blacklisted {
							return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
						}
						if err != nil {
							return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
						}
					}
				}
				for _, o := range m.Outputs {
					for _, c := range o.Coins {
						addresses := []string{o.Address}
						blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, c, ad.tokenfactory, ad.fiattokenfactory, ad.stabletokenfactory)
						if blacklisted {
							return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
						}
						if err != nil {
							return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
						}
					}
				}
			case *transfertypes.MsgTransfer:
				addresses := []string{m.Sender, m.Receiver}
				blacklisted, address, err := checkForBlacklistedAddressByTokenFactory(ctx, addresses, m.Token, ad.tokenfactory, ad.fiattokenfactory, ad.stabletokenfactory)
				if blacklisted {
					return ctx, sdkerrors.Wrapf(err, "an address (%s) is blacklisted and can not send or receive tokens", address)
				}
				if err != nil {
					return ctx, sdkerrors.Wrapf(err, "error decoding address (%s)", address)
				}
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

// checkForBlacklistedAddressByTokenFactory first checks if the denom being transacted is a mintable asset from a TokenFactory,
// if it is, it checks if the addresses involved in the tx are blacklisted by that specific TokenFactory.
func checkForBlacklistedAddressByTokenFactory(ctx sdk.Context, addresses []string, c sdk.Coin, tf *tokenfactory.Keeper, ctf *fiattokenfactory.Keeper, stf *stabletokenfactorykeeper.Keeper) (blacklisted bool, blacklistedAddress string, err error) {
	tfMintingDenom := tf.GetMintingDenom(ctx)
	if c.Denom == tfMintingDenom.Denom {
		for _, address := range addresses {
			_, addressBz, err := bech32.DecodeAndConvert(address)
			if err != nil {
				return false, address, err
			}
			_, found := tf.GetBlacklisted(ctx, addressBz)
			if found {
				return true, address, tokenfactorytypes.ErrUnauthorized
			}
		}
	}
	ctfMintingDenom := ctf.GetMintingDenom(ctx)
	if c.Denom == ctfMintingDenom.Denom {
		for _, address := range addresses {
			_, addressBz, err := bech32.DecodeAndConvert(address)
			if err != nil {
				return false, address, err
			}
			_, found := ctf.GetBlacklisted(ctx, addressBz)
			if found {
				return true, address, fiattokenfactorytypes.ErrUnauthorized
			}
		}
	}
	stfMintingDenom := stf.GetMintingDenom(ctx)
	if c.Denom == stfMintingDenom.Denom {
		for _, address := range addresses {
			_, addressBz, err := bech32.DecodeAndConvert(address)
			if err != nil {
				return false, address, err
			}
			_, found := stf.GetBlacklisted(ctx, addressBz)
			if found {
				return true, address, stabletokenfactorytypes.ErrUnauthorized
			}
		}
	}
	return false, "", nil
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
		NewIsBlacklistedDecorator(options.tokenFactoryKeeper, options.fiatTokenFactoryKeeper, options.stableTokenFactoryKeeper),
		NewIsPausedDecorator(options.tokenFactoryKeeper, options.fiatTokenFactoryKeeper, options.stableTokenFactoryKeeper),
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
