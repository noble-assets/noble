package keeper_test

import (
	"context"
	"noble/x/tokenfactory/keeper"
	"noble/x/tokenfactory/types"
	"testing"

	keepertest "noble/testutil/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.TokenfactoryKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
