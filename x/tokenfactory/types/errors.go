package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tokenfactory module sentinel errors
var (
	ErrUnauthorized = sdkerrors.Register(ModuleName, 2, "unauthorized")
	ErrUserNotFound = sdkerrors.Register(ModuleName, 3, "user not found")
)
