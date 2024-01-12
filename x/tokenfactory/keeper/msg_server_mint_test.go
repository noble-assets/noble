package keeper_test

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
)

func (suite *KeeperTestSuite) TestMint() {
	type args struct {
		fromAddr        sdk.AccAddress
		toAddr          sdk.AccAddress
		amount          sdk.Coin
		minterAddress   sdk.AccAddress
		minterAllowance sdk.Coin
	}

	type errArgs struct {
		shouldPass bool
		contains   string
	}

	tests := []struct {
		name    string
		args    args
		errArgs errArgs
	}{
		{"Valid Minter -> when the minter is authorized",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin(denom, sdk.NewInt(200)),
				minterAddress:   suite.address[0],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: true,
				contains:   "",
			},
		},
		{"Unauthorized Minter -> when the minter is not authorized (not in the list of minters)",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin(denom, sdk.NewInt(200)),
				minterAddress:   suite.address[2],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: false,
				contains:   "not a minter",
			},
		},
		{"Invalid Minting Denom -> when the Amount has a different denom than the expected minting denom",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin("mrupee", sdk.NewInt(200)),
				minterAddress:   suite.address[0],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: false,
				contains:   "incorrect minting denom",
			},
		},
		{"Insufficient Minter Allowance -> when the minter's allowance is less than the minting amount",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin(denom, sdk.NewInt(200000)),
				minterAddress:   suite.address[0],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: false,
				contains:   "",
			},
		},
		{"Paused Minting -> when minting is paused",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin(denom, sdk.NewInt(200)),
				minterAddress:   suite.address[0],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: false,
				contains:   "",
			},
		},
		{"Invalid Receiver -> when the receiver's address is blacklisted",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin(denom, sdk.NewInt(200)),
				minterAddress:   suite.address[0],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: false,
				contains:   "",
			},
		},
		{"Invalid Minter Address -> when the minter's address is blacklisted",
			args{
				fromAddr:        suite.address[0],
				toAddr:          suite.address[1],
				amount:          sdk.NewCoin(denom, sdk.NewInt(200)),
				minterAddress:   suite.address[0],
				minterAllowance: sdk.NewCoin(denom, sdk.NewInt(1000)),
			},
			errArgs{
				shouldPass: false,
				contains:   "",
			},
		},
	}
	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// set minter
			minters := types.Minters{Address: tc.args.minterAddress.String(), Allowance: tc.args.minterAllowance}

			suite.tokenFactoryKeeper.SetMinters(suite.ctx, minters)

			msg := types.MsgMint{From: tc.args.fromAddr.String(), Address: tc.args.toAddr.String(), Amount: tc.args.amount}

			res, err := suite.tokenFactoryKeeper.Mint(suite.ctx, &msg)

			val, _ := suite.tokenFactoryKeeper.GetMinters(suite.ctx, tc.args.fromAddr.String())

			if tc.errArgs.shouldPass {
				suite.Require().NoError(err, tc.name)
				suite.Require().NotNil(res)
				suite.Require().NotEmpty(val)
			} else {
				suite.Require().Error(err, tc.name)
				suite.Require().Empty(val)
				suite.Require().True(strings.Contains(err.Error(), tc.errArgs.contains))
			}
		})
	}
}
