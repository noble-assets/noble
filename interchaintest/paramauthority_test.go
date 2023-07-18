package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v3"
	"github.com/strangelove-ventures/interchaintest/v3/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v3/ibc"
	"github.com/strangelove-ventures/interchaintest/v3/testreporter"
	"github.com/strangelove-ventures/noble/cmd"
	proposaltypes "github.com/strangelove-ventures/paramauthority/x/params/types/proposal"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type ParamsCase struct {
	title         string
	description   string
	newAuthority  string
	msgAuthority  string
	signer        ibc.Wallet
	shouldSucceed bool
}

func testParamsCase(t *testing.T, ctx context.Context, broadcaster *cosmos.Broadcaster, testCase ParamsCase, chainCfg ibc.ChainConfig) {
	t.Logf(
		"SIGNER: %s\nMSG AUTHORITY: %s\n",
		testCase.signer.FormattedAddress(),
		testCase.msgAuthority,
	)
	msgUpdateParams := proposaltypes.MsgUpdateParams{
		ChangeProposal: proposal.NewParameterChangeProposal(
			testCase.title,
			testCase.description,
			[]proposal.ParamChange{
				{
					Subspace: "params",
					Key:      "authority",
					Value:    fmt.Sprintf(`"%s"`, testCase.newAuthority),
				},
			}),
		Authority: testCase.msgAuthority,
	}

	wallet := cosmos.NewWallet(
		testCase.signer.KeyName(),
		testCase.signer.Address(),
		testCase.signer.Mnemonic(),
		chainCfg,
	)

	tx, err := cosmos.BroadcastTx(
		ctx,
		broadcaster,
		wallet,
		&msgUpdateParams,
	)
	require.NoError(t, err, "failed to broadcast tx")

	t.Logf("TX: %+v\n", tx)

	if testCase.shouldSucceed {
		require.Equal(t, uint32(0), tx.Code, "changing authority failed")
	} else {
		require.NotEqual(t, uint32(0), tx.Code, "changing authority succeeded when it should have failed")
	}
}

// run `make local-image`to rebuild updated binary before running test
func TestNobleParamAuthority(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	var noble *cosmos.CosmosChain
	var roles NobleRoles
	var paramauthorityWallet ibc.Wallet
	var extraWallets ExtraWallets

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
		PreGenesis: func(cc ibc.ChainConfig) error {
			val := noble.Validators[0]
			err := createTokenfactoryRoles(ctx, &roles, denomMetadataRupee, val, true)
			if err != nil {
				return err
			}
			err = createTokenfactoryRoles(ctx, &roles, denomMetadataDrachma, val, true)
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
		ModifyGenesis: modifyGenesisAll(&roles, &roles, paramauthorityWallet.FormattedAddress()),
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

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	broadcaster := cosmos.NewBroadcaster(t, noble)

	var orderedTestCases = []ParamsCase{
		{
			title:         "change authority to alice from incorrect msg authority and signer",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  extraWallets.Alice.FormattedAddress(),
			msgAuthority:  extraWallets.User.FormattedAddress(), // matches signer, but this is not the params authority.
			signer:        extraWallets.User,
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct signer but incorrect msg authority",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  extraWallets.Alice.FormattedAddress(),
			msgAuthority:  extraWallets.User.FormattedAddress(), // this is not the params authority.
			signer:        paramauthorityWallet,                 // this is the params authority, but does not match msgAuthority
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct msg authority but incorrect signer",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  extraWallets.Alice.FormattedAddress(),
			msgAuthority:  paramauthorityWallet.FormattedAddress(), // this is the params authority.
			signer:        extraWallets.User,                       // this is not the params authority.
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct signer and msg authority",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  extraWallets.Alice.FormattedAddress(),
			msgAuthority:  paramauthorityWallet.FormattedAddress(), // this is the params authority.
			signer:        paramauthorityWallet,                    // this is the params authority.
			shouldSucceed: true,
		},
		{
			title:         "change authority to user2 from prior authority",
			description:   "change params and upgrade authority to user2's address",
			newAuthority:  extraWallets.User2.FormattedAddress(),
			msgAuthority:  paramauthorityWallet.FormattedAddress(), // this account is no longer the params authority.
			signer:        paramauthorityWallet,                    // this account is no longer the params authority.
			shouldSucceed: false,
		},
		{
			title:         "change authority to user2 from new authority",
			description:   "change params and upgrade authority to user2's address",
			newAuthority:  extraWallets.User2.FormattedAddress(),
			msgAuthority:  extraWallets.Alice.FormattedAddress(), // this account is the new params authority.
			signer:        extraWallets.Alice,                    // this account is the new params authority.
			shouldSucceed: true,
		},
	}

	for _, testCase := range orderedTestCases {
		t.Run(testCase.title, func(t *testing.T) {
			testParamsCase(t, ctx, broadcaster, testCase, chainCfg)
		})
	}

	height, err := noble.Height(ctx)
	require.NoError(t, err, "failed to get noble height")

	err = noble.StopAllNodes(ctx)
	require.NoError(t, err, "failed to stop noble chain")

	state, err := noble.ExportState(ctx, int64(height))
	require.NoError(t, err, "failed to export noble state")

	var gs interface{}
	err = json.Unmarshal([]byte(state), &gs)
	require.NoError(t, err, "failed to unmarshal state export")

	authority, err := dyno.Get(gs, "app_state", "params", "params", "authority")
	require.NoError(t, err, "failed to get authority from state export")

	require.Equal(t, extraWallets.User2.FormattedAddress(), authority, "authority does not match")

}
