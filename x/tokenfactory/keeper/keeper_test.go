package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdksimapp "github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/noble/v5/app"
	nobleApp "github.com/noble-assets/noble/v5/app"
	"github.com/noble-assets/noble/v5/cmd"
	tokenFactoryKeeper "github.com/noble-assets/noble/v5/x/tokenfactory/keeper"
	"github.com/noble-assets/noble/v5/x/tokenfactory/types"
	//paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	denom = "urupee"
	acc1  = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc2  = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
	acc3  = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address().Bytes())
)

// shared setup
type KeeperTestSuite struct {
	suite.Suite

	address            []sdk.AccAddress
	cmd                cmd.App
	app                nobleApp.App
	ctx                sdk.Context
	bankKeeper         types.BankKeeper
	tokenFactoryKeeper tokenFactoryKeeper.Keeper
	msgServer          types.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.cmd = app.Setup(false)
	suite.ctx = suite.app.NewContext(false, tmproto.Header{})
	suite.tokenFactoryKeeper = *suite.app.TokenFactoryKeeper
	suite.bankKeeper = suite.app.BankKeeper
	suite.msgServer = tokenFactoryKeeper.NewMsgServerImpl(suite.app.TokenFactoryKeeper)

	for _, acc := range []sdk.AccAddress{acc1, acc2, acc3} {
		err := sdksimapp.FundAccount(
			suite.app.BankKeeper,
			suite.ctx,
			acc,
			sdk.NewCoins(
				sdk.NewCoin(denom, sdk.NewInt(1000)),
			),
		)
		if err != nil {
			panic(err)
		}
	}

	suite.address = []sdk.AccAddress{acc1, acc2, acc3}
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
