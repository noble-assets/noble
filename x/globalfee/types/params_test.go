package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestDefaultParams(t *testing.T) {
	p := DefaultParams()
	require.EqualValues(t, p.MinimumGasPrices, sdk.DecCoins{})
	require.EqualValues(t, p.BypassMinFeeMsgTypes, []string{
		"/ibc.core.client.v1.MsgUpdateClient",
		"/ibc.core.channel.v1.MsgRecvPacket",
		"/ibc.core.channel.v1.MsgAcknowledgement",
		"/ibc.applications.transfer.v1.MsgTransfer",
		"/ibc.core.channel.v1.MsgTimeout",
		"/ibc.core.channel.v1.MsgTimeoutOnClose",
		"/cosmos.params.v1beta1.MsgUpdateParams",
		"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
		"/cosmos.upgrade.v1beta1.MsgCancelUpgrade",
		"/noble.fiattokenfactory.MsgUpdateMasterMinter",
		"/noble.fiattokenfactory.MsgUpdatePauser",
		"/noble.fiattokenfactory.MsgUpdateBlacklister",
		"/noble.fiattokenfactory.MsgUpdateOwner",
		"/noble.fiattokenfactory.MsgAcceptOwner",
		"/noble.fiattokenfactory.MsgConfigureMinter",
		"/noble.fiattokenfactory.MsgRemoveMinter",
		"/noble.fiattokenfactory.MsgMint",
		"/noble.fiattokenfactory.MsgBurn",
		"/noble.fiattokenfactory.MsgBlacklist",
		"/noble.fiattokenfactory.MsgUnblacklist",
		"/noble.fiattokenfactory.MsgPause",
		"/noble.fiattokenfactory.MsgUnpause",
		"/noble.fiattokenfactory.MsgConfigureMinterController",
		"/noble.fiattokenfactory.MsgRemoveMinterController",
		"/noble.tokenfactory.MsgUpdatePauser",
		"/noble.tokenfactory.MsgUpdateBlacklister",
		"/noble.tokenfactory.MsgUpdateOwner",
		"/noble.tokenfactory.MsgAcceptOwner",
		"/noble.tokenfactory.MsgConfigureMinter",
		"/noble.tokenfactory.MsgRemoveMinter",
		"/noble.tokenfactory.MsgMint",
		"/noble.tokenfactory.MsgBurn",
		"/noble.tokenfactory.MsgBlacklist",
		"/noble.tokenfactory.MsgUnblacklist",
		"/noble.tokenfactory.MsgPause",
		"/noble.tokenfactory.MsgUnpause",
		"/noble.tokenfactory.MsgConfigureMinterController",
		"/noble.tokenfactory.MsgRemoveMinterController",
	})
}

func Test_validateParams(t *testing.T) {
	tests := map[string]struct {
		coins     interface{} // not sdk.DeCoins, but Decoins defined in glboalfee
		expectErr bool
	}{
		"DefaultParams, pass": {
			DefaultParams().MinimumGasPrices,
			false,
		},
		"DecCoins conversion fails, fail": {
			sdk.Coins{sdk.NewCoin("photon", sdk.OneInt())},
			true,
		},
		"coins amounts are zero, pass": {
			sdk.DecCoins{
				sdk.NewDecCoin("atom", sdk.ZeroInt()),
				sdk.NewDecCoin("photon", sdk.ZeroInt()),
			},
			false,
		},
		"duplicate coins denoms, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", sdk.OneInt()),
				sdk.NewDecCoin("photon", sdk.OneInt()),
			},
			true,
		},
		"coins are not sorted by denom alphabetically, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", sdk.OneInt()),
				sdk.NewDecCoin("atom", sdk.OneInt()),
			},
			true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateMinimumGasPrices(test.coins)
			if test.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
