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
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// run `make local-image`to rebuild updated binary before running test
func TestGlobalFee(t *testing.T) {
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
		extraWallets         ExtraWallets
		paramauthorityWallet ibc.Wallet
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
		Images:         nobleImageInfo,
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) (err error) {
			val := noble.Validators[0]
			err = createTokenfactoryRoles(ctx, &roles, denomMetadataRupee, val, false)
			if err != nil {
				return err
			}
			err = createTokenfactoryRoles(ctx, &roles2, denomMetadataDrachma, val, true)
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
			if err := modifyGenesisTokenfactory(g, "tokenfactory", denomMetadataRupee, &roles, false); err != nil {
				return nil, err
			}
			if err := modifyGenesisTokenfactory(g, "fiat-tokenfactory", denomMetadataDrachma, &roles2, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisParamAuthority(g, paramauthorityWallet.Address); err != nil {
				return nil, err
			}
			if err := modifyGenesisTariffDefaults(g, paramauthorityWallet.Address); err != nil {
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

	sendAmount100 := fmt.Sprintf("100%s", chainCfg.Denom)
	minGasPriceAmount := "0.00001"

	minGasPrice := minGasPriceAmount + chainCfg.Denom
	zeroGasPrice := "0.0" + chainCfg.Denom

	// send tx with zero fees with the default MinimumGasPricesParam of 0 (null) - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName, "bank", "send", extraWallets.User2.KeyName, extraWallets.Alice.Address, sendAmount100, "--gas-prices", zeroGasPrice)
	require.NoError(t, err, "failed sending transaction")

	msgUpdateParams := proposaltypes.MsgUpdateParams{
		ChangeProposal: proposal.NewParameterChangeProposal(
			"Global Fees Param Change",
			"Update global fees",
			[]proposal.ParamChange{
				{
					Subspace: "globalfee",
					Key:      "MinimumGasPricesParam",
					Value:    fmt.Sprintf(`[{"denom":"%s", "amount":"%s"}]`, chainCfg.Denom, minGasPriceAmount),
				},
			}),
		Authority: paramauthorityWallet.Address,
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)

	decoded := sdk.MustAccAddressFromBech32(paramauthorityWallet.Address)
	wallet := &ibc.Wallet{
		Address:  string(decoded),
		Mnemonic: paramauthorityWallet.Mnemonic,
		KeyName:  paramauthorityWallet.KeyName,
		CoinType: paramauthorityWallet.CoinType,
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
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName,
		"bank", "send",
		extraWallets.User2.Address, extraWallets.Alice.Address, sendAmount100,
		"--gas-prices", zeroGasPrice,
		"-b", "block",
	)
	require.Error(t, err, "tx should not have succeeded with zero fees")

	// send tx with the gas price set by MinimumGasPricesParam - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, extraWallets.User2.KeyName,
		"bank", "send",
		extraWallets.User2.Address, extraWallets.Alice.Address, sendAmount100,
		"--gas-prices", minGasPrice,
		"-b", "block",
	)
	require.NoError(t, err, "tx should have succeeded")

	// send tx with zero fees while the default MinimumGasPricesParam requires fees, but update owner msg is in the bypass min fee msgs list - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, roles.Owner.KeyName,
		"tokenfactory", "update-owner", roles.Owner2.Address,
		"--gas-prices", zeroGasPrice,
		"-b", "block",
	)
	require.NoError(t, err, "failed to execute update owner tx with zero fees")

	// send tx with zero fees while the default MinimumGasPricesParam requires fees, but accept owner msg is in the bypass min fee msgs list - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, roles.Owner2.KeyName,
		"tokenfactory", "accept-owner",
		"--gas-prices", zeroGasPrice,
		"-b", "block",
	)
	require.NoError(t, err, "failed to execute tx to accept ownership with zero fees")

}
