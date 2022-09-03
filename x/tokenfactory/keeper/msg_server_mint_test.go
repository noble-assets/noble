package keeper_test

import (
	"testing"

	"noble/testutil/sample"

	"noble/x/tokenfactory/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgMint(t *testing.T) {
	var (
		// sdkCtx, _, ts = testkeeper.NewTestSetup(t)
		// ctx           = sdk.WrapSDKContext(sdkCtx)
		minter   = sample.Address(r)
		reciever = sample.Address(r)
		amount   = sdk.NewCoin("usdc", sdkmath.NewInt(10000))
	)
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
			// ts.TokenfactorySrv.Mint(ctx, &tc.msg)
			// Implement tests
		})
	}
}
