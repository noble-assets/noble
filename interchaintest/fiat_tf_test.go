package interchaintest_test

import (
	"context"
	"encoding/json"
	"testing"

	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestFiatTFOwner(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	gw := nobleSpinUp(ctx, t)
	noble := gw.chain
	val := noble.Validators[0]

	// -- Update Owner --

	// Update owner while paused
	pauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser.KeyName())

	w := interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-1", 1, noble)
	newOwner1 := w[0]

	_, err := val.ExecTx(ctx, gw.fiatTfRoles.Owner.KeyName(), "fiat-tokenfactory", "update-owner", string(newOwner1.FormattedAddress()))
	require.NoError(t, err, "error broadcasting update owner message")

	unpauseFiatTF(t, ctx, val, gw.fiatTfRoles.Pauser.KeyName())

	// Update owner from unprivileged account
	w = interchaintest.GetAndFundTestUsers(t, ctx, "new-owner-2", 1, noble)
	newOwner2 := w[0]

	hash, err := val.ExecTx(ctx, gw.extraWallets.Alice.KeyName(), "fiat-tokenfactory", "update-owner", string(newOwner2.FormattedAddress()))
	require.NoError(t, err, "error broadcasting update owner message")

	res, _, err := val.ExecQuery(ctx, "tx", hash)
	require.NoError(t, err, "error querying for tx hash")

	// TODO: after updating to latest interchaintest, use noble.GetTransaction(hash)
	var txResponse sdktypes.TxResponse
	_ = json.Unmarshal(res, &txResponse)
	// ignore the error since some types do not unmarshal (ex: height of int64 vs string)

	require.Contains(t, txResponse.RawLog, "you are not the owner: unauthorized")
	require.Greater(t, txResponse.Code, uint32(1))

	// Update Owner from blacklisted account

}

func pauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser string) {
	_, err := val.ExecTx(ctx, pauser, "fiat-tokenfactory", "pause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")

	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err)

	expectedPaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: true,
		},
	}
	require.Equal(t, expectedPaused, showPausedResponse)
}

func unpauseFiatTF(t *testing.T, ctx context.Context, val *cosmos.ChainNode, pauser string) {
	_, err := val.ExecTx(ctx, pauser, "fiat-tokenfactory", "unpause")
	require.NoError(t, err, "error pausing fiat-tokenfactory")

	res, _, err := val.ExecQuery(ctx, "fiat-tokenfactory", "show-paused")
	require.NoError(t, err, "error querying for paused state")

	var showPausedResponse fiattokenfactorytypes.QueryGetPausedResponse
	err = json.Unmarshal(res, &showPausedResponse)
	require.NoError(t, err, "failed to unmarshall show-puased response")

	expectedUnpaused := fiattokenfactorytypes.QueryGetPausedResponse{
		Paused: fiattokenfactorytypes.Paused{
			Paused: false,
		},
	}
	require.Equal(t, expectedUnpaused, showPausedResponse)
}

// starts noble chain and sets up all Fiat Token Factory Roles
func nobleSpinUp(ctx context.Context, t *testing.T) (gw genesisWrapper) {
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", nv, nf, true, false, true, false),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain

	// cmd.SetPrefixes(noble.Config().Bech32Prefix)

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	return
}
