package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/tokenfactory module sentinel errors
var (
	ErrMint               = sdkerrors.Register(ModuleName, 2, "tokens can't be minted")
	ErrParseAddress       = sdkerrors.Register(ModuleName, 3, "can't parse address")
	ErrSendCoinsToAccount = sdkerrors.Register(ModuleName, 4, "can't send tokens to account")
)
