package e2e_test

import (
	"context"
	"testing"

	"github.com/noble-assets/noble/e2e"
	"github.com/strangelove-ventures/interchaintest/v8/conformance"
)

func TestConformance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	ctx := context.Background()

	var nw e2e.NobleWrapper
	nw, ibcSimd, rf, r, ibcPathName, rep, _, client, network := e2e.NobleSpinUpIBC(t, ctx, false)

	conformance.TestChainPair(t, ctx, client, network, nw.Chain, ibcSimd, rf, rep, r, ibcPathName)
}
