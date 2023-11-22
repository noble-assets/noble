package interchaintest_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/noble-assets/noble/v5/cmd"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/cmd"
=======
>>>>>>> a4ad980 (chore: rename module path (#283))
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

	var gw genesisWrapper

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", 2, 0, false, true, false, true),
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	gw.chain = chains[0].(*cosmos.CosmosChain)
	noble := gw.chain

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

	chainCfg := noble.Config()

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	nobleValidator := noble.Validators[0]

	sendAmount100 := fmt.Sprintf("100%s", chainCfg.Denom)
	minGasPriceAmount := "0.00001"

	minGasPrice := minGasPriceAmount + chainCfg.Denom
	zeroGasPrice := "0.0" + chainCfg.Denom

	// send tx with zero fees with the default MinimumGasPricesParam of 0 (null) - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, gw.extraWallets.User2.KeyName(), "bank", "send", gw.extraWallets.User2.KeyName(), gw.extraWallets.Alice.FormattedAddress(), sendAmount100, "--gas-prices", zeroGasPrice)
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
		Authority: gw.paramAuthority.FormattedAddress(),
	}

	broadcaster := cosmos.NewBroadcaster(t, noble)

	wallet := cosmos.NewWallet(
		gw.paramAuthority.KeyName(),
		gw.paramAuthority.Address(),
		gw.paramAuthority.Mnemonic(),
		chainCfg,
	)

	tx, err := cosmos.BroadcastTx(
		ctx,
		broadcaster,
		wallet,
		&msgUpdateParams,
	)
	require.NoError(t, err, "failed to broadcast tx")
	require.Equal(t, uint32(0), tx.Code, "tx proposal failed")

	// send tx with zero fees while the default MinimumGasPricesParam requires fees - tx should fail
	_, err = nobleValidator.ExecTx(ctx, gw.extraWallets.User2.KeyName(),
		"bank", "send",
		gw.extraWallets.User2.FormattedAddress(), gw.extraWallets.Alice.FormattedAddress(), sendAmount100,
		"--gas-prices", zeroGasPrice,
		"-b", "block",
	)
	require.Error(t, err, "tx should not have succeeded with zero fees")

	// send tx with the gas price set by MinimumGasPricesParam - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, gw.extraWallets.User2.KeyName(),
		"bank", "send",
		gw.extraWallets.User2.FormattedAddress(), gw.extraWallets.Alice.FormattedAddress(), sendAmount100,
		"--gas-prices", minGasPrice,
		"-b", "block",
	)
	require.NoError(t, err, "tx should have succeeded")

	// send tx with zero fees while the default MinimumGasPricesParam requires fees, but update owner msg is in the bypass min fee msgs list - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, gw.tfRoles.Owner.KeyName(),
		"tokenfactory", "update-owner", gw.tfRoles.Owner2.FormattedAddress(),
		"--gas-prices", zeroGasPrice,
		"-b", "block",
	)
	require.NoError(t, err, "failed to execute update owner tx with zero fees")

	// send tx with zero fees while the default MinimumGasPricesParam requires fees, but accept owner msg is in the bypass min fee msgs list - tx should succeed
	_, err = nobleValidator.ExecTx(ctx, gw.tfRoles.Owner2.KeyName(),
		"tokenfactory", "accept-owner",
		"--gas-prices", zeroGasPrice,
		"-b", "block",
	)
	require.NoError(t, err, "failed to execute tx to accept ownership with zero fees")

}
