package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/x/tokenfactory/keeper"
	"github.com/strangelove-ventures/noble/x/tokenfactory/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.TokenfactoryKeeper(t)
	return keeper.NewMsgServerImpl(k), sdk.WrapSDKContext(ctx)
}
