package ibctest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/ibctest/v3"
	"github.com/strangelove-ventures/ibctest/v3/chain/cosmos"
	"github.com/strangelove-ventures/ibctest/v3/ibc"
	"github.com/strangelove-ventures/ibctest/v3/testreporter"
	"github.com/strangelove-ventures/noble/cmd"
	integration "github.com/strangelove-ventures/noble/ibctest"
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

func testParamsCase(t *testing.T, ctx context.Context, broadcaster *cosmos.Broadcaster, testCase ParamsCase) {
	t.Logf(
		"SIGNER: %s\nMSG AUTHORITY: %s\n",
		testCase.signer.Address,
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

	decoded := sdk.MustAccAddressFromBech32(testCase.signer.Address)
	wallet := &ibc.Wallet{
		Address:  string(decoded),
		Mnemonic: testCase.signer.Mnemonic,
		KeyName:  testCase.signer.KeyName,
		CoinType: testCase.signer.CoinType,
	}

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

func TestNobleParamAuthority(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := ibctest.DockerSetup(t)

	repo, version := integration.GetDockerImageInfo()

	var noble *cosmos.CosmosChain
	var roles NobleRoles

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
		Images: []ibc.DockerImage{
			{
				Repository: repo,
				Version:    version,
				UidGid:     "1025:1025",
			},
		},
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) error {
			val := noble.Validators[0]
			_, _, err := val.ExecBin(ctx, "add-consumer-section")
			if err != nil {
				return err
			}
			roles, err = noblePreGenesis(ctx, val)
			if err != nil {
				return err
			}
			return nil
		},
		ModifyGenesis: func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
			return modifyGenesisNobleOwner(b, roles.Owner.Address)
		},
	}

	nv := 1
	nf := 0

	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*ibctest.ChainSpec{
		{
			ChainConfig:   chainCfg,
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	noble = chains[0].(*cosmos.CosmosChain)

	ic := ibctest.NewInterchain().
		AddChain(noble)

	require.NoError(t, ic.Build(ctx, eRep, ibctest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),

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
			newAuthority:  roles.Alice.Address,
			msgAuthority:  roles.User.Address, // matches signer, but this is not the params authority.
			signer:        roles.User,
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct signer but incorrect msg authority",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  roles.Alice.Address,
			msgAuthority:  roles.User.Address, // this is not the params authority.
			signer:        roles.Owner,        // this is the params authority, but does not match msgAuthority
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct msg authority but incorrect signer",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  roles.Alice.Address,
			msgAuthority:  roles.Owner.Address, // this is the params authority.
			signer:        roles.User,          // this is not the params authority.
			shouldSucceed: false,
		},
		{
			title:         "change authority to alice from correct signer and msg authority",
			description:   "change params and upgrade authority to alice's address",
			newAuthority:  roles.Alice.Address,
			msgAuthority:  roles.Owner.Address, // this is the params authority.
			signer:        roles.Owner,         // this is the params authority.
			shouldSucceed: true,
		},
		{
			title:         "change authority to user2 from prior authority",
			description:   "change params and upgrade authority to user2's address",
			newAuthority:  roles.User2.Address,
			msgAuthority:  roles.Owner.Address, // this account is no longer the params authority.
			signer:        roles.Owner,         // this account is no longer the params authority.
			shouldSucceed: false,
		},
		{
			title:         "change authority to user2 from new authority",
			description:   "change params and upgrade authority to user2's address",
			newAuthority:  roles.User2.Address,
			msgAuthority:  roles.Alice.Address, // this account is the new params authority.
			signer:        roles.Alice,         // this account is the new params authority.
			shouldSucceed: true,
		},
	}

	for _, testCase := range orderedTestCases {
		t.Run(testCase.title, func(t *testing.T) {
			testParamsCase(t, ctx, broadcaster, testCase)
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

	require.Equal(t, roles.User2.Address, authority, "authority does not match")

}
