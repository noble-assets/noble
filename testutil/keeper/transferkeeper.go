package keeper

import (
	"context"

	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
)

type MockTransferKeeper struct{}

func (MockTransferKeeper) Transfer(ctx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	return &types.MsgTransferResponse{Sequence: 0}, nil
}
