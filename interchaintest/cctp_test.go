package interchaintest_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/noble/cmd"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	cctptypes "github.com/strangelove-ventures/noble/x/cctp/types"
)

// run `make local-image`to rebuild updated binary before running test
func TestCCTP(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var gw genesisWrapper

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", nv, nf, true, true, true, true),
		{
			Name:          "gaia",
			Version:       "v10.0.2",
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain
	gaia := chains[1].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		relayerImage,
	).Build(t, client, network)

	pathName := "noble-gaia"

	ic := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(r, "r").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: r,
			Path:    pathName,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	nobleChainCfg := noble.Config()

	cmd.SetPrefixes(nobleChainCfg.Bech32Prefix)

	attesters := make([]*ecdsa.PrivateKey, 3)
	msgs := make([]sdk.Msg, 3)

	for i := range attesters {
		p, err := crypto.GenerateKey()
		require.NoError(t, err)

		attesters[i] = p

		pubKey := elliptic.Marshal(p.PublicKey, p.PublicKey.X, p.PublicKey.Y)

		msgs[i] = &cctptypes.MsgAddPublicKey{
			From:      gw.fiatTfRoles.Owner.FormattedAddress(),
			PublicKey: pubKey,
		}
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		msgs...,
	)
	require.NoError(t, err, "error submitting add public keys tx")

	mockMessaage := []byte("hello world") // TODO

	digest := crypto.Keccak256(mockMessaage)

	attestation := make([]byte, 0, len(attesters)*65)

	// CCTP requires attestations to have signatures sorted by address
	sort.Slice(attesters, func(i, j int) bool {
		return bytes.Compare(
			crypto.PubkeyToAddress(attesters[i].PublicKey).Bytes(),
			crypto.PubkeyToAddress(attesters[j].PublicKey).Bytes(),
		) < 0
	})

	for i := range attesters {
		sig, err := crypto.Sign(digest, attesters[i])
		require.NoError(t, err)

		attestation = append(attestation, sig...)
	}

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		&cctptypes.MsgReceiveMessage{
			From:        gw.fiatTfRoles.Owner.FormattedAddress(),
			Message:     mockMessaage,
			Attestation: attestation,
		},
	)
	require.NoError(t, err, "error submitting cctp recv tx")

}
