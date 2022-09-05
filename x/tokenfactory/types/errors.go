package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tokenfactory module sentinel errors
var (
	ErrUnauthorized       = sdkerrors.Register(ModuleName, 2, "unauthorized")
	ErrUserNotFound       = sdkerrors.Register(ModuleName, 3, "user not found")
	ErrMint               = sdkerrors.Register(ModuleName, 4, "tokens can not be minted")
	ErrParseAddress       = sdkerrors.Register(ModuleName, 5, "can't parse address")
	ErrSendCoinsToAccount = sdkerrors.Register(ModuleName, 6, "can't send tokens to account")
)
