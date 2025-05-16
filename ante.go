// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package noble

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory"
	ftfkeeper "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/noble-assets/forwarding/v2"
	forwardingtypes "github.com/noble-assets/forwarding/v2/types"
)

type BankKeeper interface {
	authtypes.BankKeeper
	forwardingtypes.BankKeeper
}

// HandlerOptions extends the options required by the default Cosmos SDK
// AnteHandler for our custom ante decorators.
type HandlerOptions struct {
	ante.HandlerOptions
	cdc        codec.Codec
	BankKeeper BankKeeper
	FTFKeeper  *ftfkeeper.Keeper
	IBCKeeper  *ibckeeper.Keeper
}

// NewAnteHandler extends the default Cosmos SDK AnteHandler with custom ante decorators.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.FTFKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "fiattokenfactory keeper is required for ante builder")
	}

	if options.IBCKeeper == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "ibc keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),

		fiattokenfactory.NewIsPausedDecorator(options.cdc, options.FTFKeeper),
		fiattokenfactory.NewIsBlacklistedDecorator(options.FTFKeeper),

		NewPermissionedHyperlaneDecorator(),
		NewPermissionedLiquidityDecorator(),

		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),

		NewSigVerificationDecorator(options),

		ante.NewIncrementSequenceDecorator(options.AccountKeeper),

		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

func NewSigVerificationDecorator(options HandlerOptions) sdk.AnteDecorator {
	defaultAnte := ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler)
	return forwarding.NewSigVerificationDecorator(options.BankKeeper, defaultAnte)
}

// SigVerificationGasConsumer is a custom implementation of the signature verification gas
// consumer to handle public keys defined in the Forwarding module.
func SigVerificationGasConsumer(meter storetypes.GasMeter, sig signing.SignatureV2, params authtypes.Params) error {
	switch sig.PubKey.(type) {
	case *forwardingtypes.ForwardingPubKey:
		return nil
	default:
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}
