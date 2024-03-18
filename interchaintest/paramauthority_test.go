package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/icza/dyno"
	"github.com/noble-assets/noble/v5/cmd"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
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

	var gw genesisWrapper

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &gw, "noble-1", 1, 0, true, true, true, true),
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

	broadcaster := cosmos.NewBroadcaster(t, noble)

	orderedTestCases := []ParamsCase{
		{
			title:         "change authority to alice from incorrect msg authority and signer",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  gw.extraWallets.Alice.FormattedAddress(),
			msgAuthority:  gw.extraWallets.User.FormattedAddress(), // matches signer, but this is not the params authority.
			signer:        gw.extraWallets.User,
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct signer but incorrect msg authority",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  gw.extraWallets.Alice.FormattedAddress(),
			msgAuthority:  gw.extraWallets.User.FormattedAddress(), // this is not the params authority.
			signer:        gw.paramAuthority,                       // this is the params authority, but does not match msgAuthority
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct msg authority but incorrect signer",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  gw.extraWallets.Alice.FormattedAddress(),
			msgAuthority:  gw.paramAuthority.FormattedAddress(), // this is the params authority.
			signer:        gw.extraWallets.User,                 // this is not the params authority.
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct signer and msg authority",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  gw.extraWallets.Alice.FormattedAddress(),
			msgAuthority:  gw.paramAuthority.FormattedAddress(), // this is the params authority.
			signer:        gw.paramAuthority,                    // this is the params authority.
			shouldSucceed: true,
		},
		{
			title:         "change authority to user2 from prior authority",
			description:   "change params and upgrade authority to user2's address",
			newAuthority:  gw.extraWallets.User2.FormattedAddress(),
			msgAuthority:  gw.paramAuthority.FormattedAddress(), // this account is no longer the params authority.
			signer:        gw.paramAuthority,                    // this account is no longer the params authority.
			shouldSucceed: false,
		},
		{
			title:         "change authority to user2 from new authority",
			description:   "change params and upgrade authority to user2's address",
			newAuthority:  gw.extraWallets.User2.FormattedAddress(),
			msgAuthority:  gw.extraWallets.Alice.FormattedAddress(), // this account is the new params authority.
			signer:        gw.extraWallets.Alice,                    // this account is the new params authority.
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

	require.Equal(t, gw.extraWallets.User2.FormattedAddress(), authority, "authority does not match")
}
