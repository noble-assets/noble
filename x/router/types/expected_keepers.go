package types

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	cctptypes "github.com/strangelove-ventures/noble-cctp/x/cctp/types"
)

// TransferKeeper defines the expected transfer keeper
type TransferKeeper interface {
	Transfer(ctx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error)
}

// CctpKeeper defines the expected cctp keeper
type CctpKeeper interface {
	GetTokenPair(ctx sdk.Context, remoteDomain uint32, remoteToken string) (val cctptypes.TokenPairs, found bool)
}
