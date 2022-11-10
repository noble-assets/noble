package types

import sdkerrors "cosmossdk.io/errors"

var (
	ErrInvalidSwapMetadata = sdkerrors.Register(ModuleName, 2, "invalid swap metadata")
	ErrSwapFailed          = sdkerrors.Register(ModuleName, 3, "swap failed")
)
