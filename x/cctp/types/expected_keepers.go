package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
	routertypes "github.com/strangelove-ventures/noble/x/router/types"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	// Methods imported from account should be defined here
}

type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

type FiatTokenfactoryKeeper interface {
	Burn(ctx context.Context, msg *fiattokenfactorytypes.MsgBurn) (*types.MsgBurnResponse, error)
	Mint(ctx context.Context, msg *fiattokenfactorytypes.MsgMint) (*types.MsgMintResponse, error)
}

type RouterKeeper interface {
	HandleMessage(ctx sdk.Context, msg *routertypes.MsgHandleMessage) (*routertypes.MsgHandleMessageResponse, error)
	GetDenom(ctx sdk.Context, key string) (val routertypes.Denoms, found bool)
}
