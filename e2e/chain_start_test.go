package e2e_test

import (
	"context"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNobleStart(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	var gw genesisWrapper

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", 2, 0, false, false),
		{Name: "gaia", Version: "latest"},
	})

	var ibcSimApp ibc.Chain

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	gw.chain, ibcSimApp = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)
	noble := gw.chain

	client, network := interchaintest.DockerSetup(t)
	rf := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t))
	r := rf.Build(t, client, network)

	const ibcPath = "path"
	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(ibcSimApp).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  ibcSimApp,
			Relayer: r,
			Path:    ibcPath,
		})

	rep := testreporter.NewNopReporter()

	require.NoError(t, ic.Build(ctx, rep.RelayerExecReporter(t), interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	conformance.TestChainPair(t, ctx, client, network, noble, ibcSimApp, rf, rep, r, ibcPath)
}
