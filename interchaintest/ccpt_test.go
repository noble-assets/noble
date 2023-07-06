package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"crypto/ecdsa"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	cctp "github.com/strangelove-ventures/noble-cctp"
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

	var (
		noble                *cosmos.CosmosChain
		roles                NobleRoles
		roles2               NobleRoles
		paramauthorityWallet ibc.Wallet
	)

	// create secp256k1 key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println("Failed to generate private key:", err)
		return
	}

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "token",
		Bech32Prefix:   "noble",
		CoinType:       "118",
		GasPrices:      "0.0token",
		GasAdjustment:  1.1,
		TrustingPeriod: "504h",
		NoHostMount:    false,
		Images:         nobleImageInfo,
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) (err error) {
			val := noble.Validators[0]
			err = createTokenfactoryRoles(ctx, &roles, denomMetadataRupee, val, false)
			if err != nil {
				return err
			}
			err = createTokenfactoryRoles(ctx, &roles2, denomMetadataDrachma, val, false)
			if err != nil {
				return err
			}
			paramauthorityWallet, err = createParamAuthAtGenesis(ctx, val)
			return err
		},
		ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			g := make(map[string]interface{})
			if err := json.Unmarshal(b, &g); err != nil {
				return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
			}
			if err := modifyGenesisTokenfactory(g, "tokenfactory", denomMetadataRupee, &roles, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisTokenfactory(g, "fiat-tokenfactory", denomMetadataDrachma, &roles2, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisParamAuthority(g, paramauthorityWallet.FormattedAddress()); err != nil {
				return nil, err
			}
			if err := modifyGenesisTariffDefaults(g, paramauthorityWallet.FormattedAddress()); err != nil {
				return nil, err
			}

			// TODO load secp256k1 pubkey into CCTP genesis state for attester.
			// publicKey := privateKey.PublicKey

			out, err := json.Marshal(&g)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
			}
			return out, nil
		},
	}

	nv := 1
	nf := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			ChainConfig:   chainCfg,
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	noble = chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	// TODO:

	// assemble initial CCTP message to mint USDC
	
	cctp.
	// take keccak256 hash of message

	// sign hash with secp256k1 private key

	// assemble noble "MsgReceiveCCTP" message

	// send message to noble

	// confirm USDC was minted

	// next, error cases
	// 1. generate different secp256k1 key, make sure error to mint
	// 2. same key, but sign invalid hash, make sure error to mint

	// next, IBC forwarding. Add gaia, relayer, etc. include forwarding instructions in initial CCTP message. make sure forwarded to gaia receiever.
}
