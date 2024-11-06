package e2e_test

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	globalfeetypes "github.com/noble-assets/globalfee/types"
	"github.com/noble-assets/noble/e2e"
	"github.com/noble-assets/noble/upgrade"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
)

func TestChainUpgrade(t *testing.T) {

	genesisVersion := "v8.0.0-rc.3"

	upgrades := []e2e.ChainUpgrade{
		{
			Image:       e2e.LocalImages[0],
			UpgradeName: "v8.0.0-rc.4",
			Emergency:   false,
			PreUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet) {
				require.False(t, noble.GetNode().HasCommand(ctx, "query", "globalfee"))
			},
			PostUpgrade: func(t *testing.T, ctx context.Context, noble *cosmos.CosmosChain, authority ibc.Wallet) {
				require.True(t, noble.GetNode().HasCommand(ctx, "query", "globalfee"))

				val := noble.Validators[0]

				bypassMessages := []string{
					sdk.MsgTypeURL(&clienttypes.MsgUpdateClient{}),
					sdk.MsgTypeURL(&channeltypes.MsgRecvPacket{}),
					sdk.MsgTypeURL(&channeltypes.MsgTimeout{}),
					sdk.MsgTypeURL(&channeltypes.MsgAcknowledgement{}),
				}
				registry := noble.Config().EncodingConfig.InterfaceRegistry
				bypassMessages = append(bypassMessages, upgrade.GetModuleMessages(registry, "circle")...)
				bypassMessages = append(bypassMessages, upgrade.GetModuleMessages(registry, "aura")...)
				bypassMessages = append(bypassMessages, upgrade.GetModuleMessages(registry, "halo")...)
				bypassMessages = append(bypassMessages, upgrade.GetModuleMessages(registry, "florin")...)

				res, _, err := val.ExecQuery(ctx, "globalfee", "bypass-messages")
				require.NoError(t, err)

				var bypassMessagesRes globalfeetypes.QueryBypassMessagesResponse
				err = json.Unmarshal(res, &bypassMessagesRes)
				require.NoError(t, err)
				require.ElementsMatch(t, bypassMessages, bypassMessagesRes.BypassMessages)
			},
		},
	}

	e2e.TestChainUpgrade(t, genesisVersion, upgrades)

}
