package interchaintest_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
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
	attesterPubs := make([][]byte, 3)

	curve := secp256k1.S256()

	for i := range attesters {
		p, err := ecdsa.GenerateKey(curve, rand.Reader)
		require.NoError(t, err)

		attesters[i] = p

		pubKey := elliptic.Marshal(p.PublicKey, p.PublicKey.X, p.PublicKey.Y)

		attesterPubs[i] = pubKey
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		gw.fiatTfRoles.Owner,
		&cctptypes.MsgAddPublicKey{
			From:      gw.fiatTfRoles.Owner.FormattedAddress(),
			PublicKey: attesterPubs[0],
		},
		&cctptypes.MsgAddPublicKey{
			From:      gw.fiatTfRoles.Owner.FormattedAddress(),
			PublicKey: attesterPubs[1],
		},
		&cctptypes.MsgAddPublicKey{
			From:      gw.fiatTfRoles.Owner.FormattedAddress(),
			PublicKey: attesterPubs[2],
		},
	)
	require.NoError(t, err, "error submitting add public keys tx")

	mockMessaage := []byte("hello world") // TODO

	digest := crypto.Keccak256(mockMessaage)

	var attestation []byte

	for i := range attesters {
		r, s, err := ecdsa.Sign(rand.Reader, attesters[i], digest)
		require.NoError(t, err)

		attestation = append(attestation, r.Bytes()...)
		attestation = append(attestation, s.Bytes()...)
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
