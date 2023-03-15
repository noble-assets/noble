package app

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
	tokenfactory "github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"
	tokenfactory_1 "github.com/strangelove-ventures/noble/x/tokenfactory1/keeper"
	tokenfactory_1types "github.com/strangelove-ventures/noble/x/tokenfactory1/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibcante "github.com/cosmos/ibc-go/v3/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
)

type HandlerOptions struct {
	ante.HandlerOptions
	tokenFactoryKeeper   *tokenfactory.Keeper
	tokenFactory_1Keeper *tokenfactory_1.Keeper
	IBCKeeper            *ibckeeper.Keeper
}

type IsPausedDecorator struct {
	tokenfactory   *tokenfactory.Keeper
	tokenfactory_1 *tokenfactory_1.Keeper
}

func NewIsPausedDecorator(tk *tokenfactory.Keeper, tk_1 *tokenfactory_1.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		tokenfactory:   tk,
		tokenfactory_1: tk_1,
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
					paused, err := checkPausedStatebyTokenfactory(ctx, c, ad.tokenfactory, ad.tokenfactory_1)
					if paused {
						return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						paused, err := checkPausedStatebyTokenfactory(ctx, c, ad.tokenfactory, ad.tokenfactory_1)
						if paused {
							return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
						}
					}
				}
			case *transfertypes.MsgTransfer:
				paused, err := checkPausedStatebyTokenfactory(ctx, m.Token, ad.tokenfactory, ad.tokenfactory_1)
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

func checkPausedStatebyTokenfactory(ctx sdk.Context, c sdk.Coin, tk *tokenfactory.Keeper, tk_1 *tokenfactory_1.Keeper) (bool, *sdkerrors.Error) {
	mintingDenom := tk.GetMintingDenom(ctx)
	if c.Denom == mintingDenom.Denom {
		paused := tk.GetPaused(ctx)
		if paused.Paused {
			return true, tokenfactorytypes.ErrPaused
		}
	}
	mintingDenom_1 := tk_1.GetMintingDenom(ctx)
	if c.Denom == mintingDenom_1.Denom {
		paused := tk_1.GetPaused(ctx)
		if paused.Paused {
			return true, tokenfactory_1types.ErrPaused
		}
	}
	return false, nil
}

type IsBlacklistedDecorator struct {
	tokenfactory   *tokenfactory.Keeper
	tokenfactory_1 *tokenfactory_1.Keeper
}

func NewIsBlacklistedDecorator(tk *tokenfactory.Keeper, tk_1 *tokenfactory_1.Keeper) IsBlacklistedDecorator {
	return IsBlacklistedDecorator{
		tokenfactory:   tk,
		tokenfactory_1: tk_1,
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
				for _, c := range m.Amount {
					if checkIfMintedAsset(ctx, c.Denom, ad.tokenfactory, ad.tokenfactory_1) {
						addresses = append(addresses, m.ToAddress, m.FromAddress)
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						if checkIfMintedAsset(ctx, c.Denom, ad.tokenfactory, ad.tokenfactory_1) {
							addresses = append(addresses, i.Address)
						}
					}
				}
				for _, o := range m.Outputs {
					for _, c := range o.Coins {
						if checkIfMintedAsset(ctx, c.Denom, ad.tokenfactory, ad.tokenfactory_1) {
							addresses = append(addresses, o.Address)
						}
					}
				}
			case *transfertypes.MsgTransfer:
				if checkIfMintedAsset(ctx, m.Token.Denom, ad.tokenfactory, ad.tokenfactory_1) {
					addresses = append(addresses, m.Sender, m.Receiver)
				}
			}

			for _, address := range addresses {
				_, addressBz, err := bech32.DecodeAndConvert(address)
				if err != nil {
					return ctx, err
				}

				_, found := ad.tokenfactory.GetBlacklisted(ctx, addressBz)
				if found {
					return ctx, sdkerrors.Wrapf(tokenfactorytypes.ErrUnauthorized, "an address (%s) is blacklisted and can not send or receive tokens", address)
				}
				_, found = ad.tokenfactory_1.GetBlacklisted(ctx, addressBz)
				if found {
					return ctx, sdkerrors.Wrapf(tokenfactory_1types.ErrUnauthorized, "an address (%s) is blacklisted and can not send or receive tokens", address)
				}
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

func checkIfMintedAsset(ctx sdk.Context, denom string, k *tokenfactory.Keeper, k_1 *tokenfactory_1.Keeper) bool {
	mintingDenom := k.GetMintingDenom(ctx)
	mintingDenom_1 := k_1.GetMintingDenom(ctx)
	mintingDenoms := [2]string{mintingDenom.Denom, mintingDenom_1.Denom}
	for _, v := range mintingDenoms {
		if v == denom {
			return true
		}
	}
	return false
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
		ante.NewRejectExtensionOptionsDecorator(),
		NewIsBlacklistedDecorator(options.tokenFactoryKeeper, options.tokenFactory_1Keeper),
		NewIsPausedDecorator(options.tokenFactoryKeeper, options.tokenFactory_1Keeper),
		ante.NewMempoolFeeDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
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
