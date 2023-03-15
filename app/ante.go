package app

import (
	"github.com/cosmos/cosmos-sdk/types/bech32"
	circletokenfactory "github.com/strangelove-ventures/noble/x/circletokenfactory/keeper"
	circletokenfactorytypes "github.com/strangelove-ventures/noble/x/circletokenfactory/types"
	tokenfactory "github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/strangelove-ventures/noble/x/tokenfactory/types"

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
	tokenFactoryKeeper       *tokenfactory.Keeper
	circletokenFactoryKeeper *circletokenfactory.Keeper
	IBCKeeper                *ibckeeper.Keeper
}

type IsPausedDecorator struct {
	tokenfactory       *tokenfactory.Keeper
	circletokenfactory *circletokenfactory.Keeper
}

func NewIsPausedDecorator(tf *tokenfactory.Keeper, ctf *circletokenfactory.Keeper) IsPausedDecorator {
	return IsPausedDecorator{
		tokenfactory:       tf,
		circletokenfactory: ctf,
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
					paused, err := checkPausedStatebyTokenfactory(ctx, c, ad.tokenfactory, ad.circletokenfactory)
					if paused {
						return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						paused, err := checkPausedStatebyTokenfactory(ctx, c, ad.tokenfactory, ad.circletokenfactory)
						if paused {
							return ctx, sdkerrors.Wrapf(err, "can not perform token transfers")
						}
					}
				}
			case *transfertypes.MsgTransfer:
				paused, err := checkPausedStatebyTokenfactory(ctx, m.Token, ad.tokenfactory, ad.circletokenfactory)
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

func checkPausedStatebyTokenfactory(ctx sdk.Context, c sdk.Coin, tf *tokenfactory.Keeper, ctf *circletokenfactory.Keeper) (bool, *sdkerrors.Error) {
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
			return true, circletokenfactorytypes.ErrPaused
		}
	}
	return false, nil
}

type IsBlacklistedDecorator struct {
	tokenfactory       *tokenfactory.Keeper
	circletokenfactory *circletokenfactory.Keeper
}

func NewIsBlacklistedDecorator(tf *tokenfactory.Keeper, ctf *circletokenfactory.Keeper) IsBlacklistedDecorator {
	return IsBlacklistedDecorator{
		tokenfactory:       tf,
		circletokenfactory: ctf,
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
					if checkIfMintedAsset(ctx, c.Denom, ad.tokenfactory, ad.circletokenfactory) {
						addresses = append(addresses, m.ToAddress, m.FromAddress)
					}
				}
			case *banktypes.MsgMultiSend:
				for _, i := range m.Inputs {
					for _, c := range i.Coins {
						if checkIfMintedAsset(ctx, c.Denom, ad.tokenfactory, ad.circletokenfactory) {
							addresses = append(addresses, i.Address)
						}
					}
				}
				for _, o := range m.Outputs {
					for _, c := range o.Coins {
						if checkIfMintedAsset(ctx, c.Denom, ad.tokenfactory, ad.circletokenfactory) {
							addresses = append(addresses, o.Address)
						}
					}
				}
			case *transfertypes.MsgTransfer:
				if checkIfMintedAsset(ctx, m.Token.Denom, ad.tokenfactory, ad.circletokenfactory) {
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
				_, found = ad.circletokenfactory.GetBlacklisted(ctx, addressBz)
				if found {
					return ctx, sdkerrors.Wrapf(circletokenfactorytypes.ErrUnauthorized, "an address (%s) is blacklisted and can not send or receive tokens", address)
				}
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}

func checkIfMintedAsset(ctx sdk.Context, denom string, tfk *tokenfactory.Keeper, ctfk *circletokenfactory.Keeper) bool {
	tfMintingDenom := tfk.GetMintingDenom(ctx)
	ctfMintingDenom := ctfk.GetMintingDenom(ctx)
	mintingDenoms := [2]string{tfMintingDenom.Denom, ctfMintingDenom.Denom}
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
		NewIsBlacklistedDecorator(options.tokenFactoryKeeper, options.circletokenFactoryKeeper),
		NewIsPausedDecorator(options.tokenFactoryKeeper, options.circletokenFactoryKeeper),
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
