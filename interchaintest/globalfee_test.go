package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/noble/cmd"
	integration "github.com/strangelove-ventures/noble/interchaintest"
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGlobalFee(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	repo, version := integration.GetDockerImageInfo()

	var (
		noble                *cosmos.CosmosChain
		roles                NobleRoles
		roles2               NobleRoles
		extraWallets         ExtraWallets
		paramauthorityWallet Authority
	)

	chainCfg := ibc.ChainConfig{
		Type:           "cosmos",
		Name:           "noble",
		ChainID:        "noble-1",
		Bin:            "nobled",
		Denom:          "utoken",
		Bech32Prefix:   "noble",
		CoinType:       "118",
		GasPrices:      "0.0utoken",
		GasAdjustment:  1.1,
		TrustingPeriod: "504h",
		NoHostMount:    false,
		Images: []ibc.DockerImage{
			{
				Repository: repo,
				Version:    version,
				UidGid:     "1025:1025",
			},
		},
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) (err error) {
			val := noble.Validators[0]
			err = createTokenfactoryRoles(ctx, &roles, DenomMetadata_rupee, val, true)
			if err != nil {
				return err
			}
			err = createTokenfactoryRoles(ctx, &roles2, DenomMetadata_drachma, val, true)
			if err != nil {
				return err
			}
			extraWallets, err = createExtraWalletsAtGenesis(ctx, val)
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
			err := modifyGenesisTokenfactory(g, "tokenfactory", DenomMetadata_rupee, &roles, true)
			if err != nil {
				return nil, err
			}
			err = modifyGenesisTokenfactory(g, "fiat-tokenfactory", DenomMetadata_drachma, &roles2, true)
			if err != nil {
				return nil, err
			}
			err = modifyGenesisParamAuthority(g, paramauthorityWallet.Authority.Address)
			if err != nil {
				return nil, err
			}
			out, err := json.Marshal(&g)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
			}
			return out, nil
		},
	}

	nv := 2
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

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	nobleValidator := noble.Validators[0]

	amount100 := fmt.Sprintf("100%s", chainCfg.Denom)
	amount0 := fmt.Sprintf("0%s", chainCfg.Denom)
	amount2 := fmt.Sprintf("2.0%s", chainCfg.Denom)

	// send tx with zero fees with the default MinimumGasPricesParam of 0 (null) - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName, "bank", "send", extraWallets.User2.KeyName, extraWallets.Alice.Address, amount100, "--fees", amount0)
	require.NoError(t, err, "failed sending transaction")

	msgUpdateParams := proposaltypes.MsgUpdateParams{
		ChangeProposal: proposal.NewParameterChangeProposal(
			"Global Fees Param Change",
			"Update global fees",
			[]proposal.ParamChange{
				{
					Subspace: "globalfee",
					Key:      "MinimumGasPricesParam",
					Value:    fmt.Sprintf(`[{"denom":"%s", "amount":"0.00001"}]`, chainCfg.Denom),
				},
			}),
		Authority: paramauthorityWallet.Authority.Address,
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)

	decoded := sdk.MustAccAddressFromBech32(paramauthorityWallet.Authority.Address)
	wallet := &ibc.Wallet{
		Address:  string(decoded),
		Mnemonic: paramauthorityWallet.Authority.Mnemonic,
		KeyName:  paramauthorityWallet.Authority.KeyName,
		CoinType: paramauthorityWallet.Authority.CoinType,
	}
	tx, err := cosmos.BroadcastTx(
		ctx,
		broadcaster,
		wallet,
		&msgUpdateParams,
	)
	require.NoError(t, err, "failed to broadcast tx")
	require.Equal(t, uint32(0), tx.Code, "tx proposal failed")

	// send tx with zero fees while the default MinimumGasPricesParam requires fees - tx should fail
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName, "bank", "send", extraWallets.User2.Address, extraWallets.Alice.Address, amount100, "--fees", amount0)
	require.Error(t, err, "tx should not have succeeded with zero fees")

	// send tx with enough fees to satisfy the MinimumGasPricesParam - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName, "bank", "send", extraWallets.User2.Address, extraWallets.Alice.Address, amount100, "--fees", amount2)
	require.NoError(t, err, "tx should have succeeded")

}
