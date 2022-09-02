package keeper_test

import (
	"context"
	"testing"

	keepertest "noble/testutil/keeper"
	"noble/testutil/sample"
	"noble/x/tokenfactory/keeper"

	_ "github.com/stretchr/testify/require"

	"noble/x/tokenfactory/types"

	testkeeper "noble/testutil/keeper"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t *testing.T) (types.MsgServer, context.Context) {
	var (
		_, _, ts = testkeeper.NewTestSetup(t)
		minter   = sample.Address(r)
		reciever = sample.Address(r)
		amount   = sdk.NewCoin("usdc", sdkmath.NewInt(10000))
	)
	k, ctx := keepertest.TokenfactoryKeeper(t)
	for _, tc := range []struct {
		name string
		msg  types.MsgMint
		err  error
	}{
		{
			name: "mint tokens",
			msg: types.MsgMint{
				From:    minter,
				Amount:  amount,
				Address: reciever,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts.TokenfactorySrv.Mint(ctx, &tc.msg)
		})
	}
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
