package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	providerTypes "github.com/cosmos/interchain-security/v2/x/ccv/provider/types"
	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"path/filepath"
	"testing"
	"time"
)

type ConsumerAdditionJSON struct {
	providerTypes.ConsumerAdditionProposal
	Deposit string `json:"deposit"`
}

func modifyGaiaGenesis(_ ibc.ChainConfig, bz []byte) ([]byte, error) {
	genesis := make(map[string]interface{})

	if err := json.Unmarshal(bz, &genesis); err != nil {
		return nil, err
	}

	err := dyno.Set(genesis, (2 * time.Minute).String(), "app_state", "gov", "voting_params", "voting_period")
	if err != nil {
		return nil, err
	}

	return json.Marshal(genesis)
}

func TestICS_StandaloneToConsumerMigration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := zaptest.NewLogger(t)
	reporter := testreporter.NewNopReporter()
	execReporter := reporter.RelayerExecReporter(t)
	client, network := interchaintest.DockerSetup(t)

	var wrapper genesisWrapper

	factory := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
		nobleChainSpec(ctx, &wrapper, "grand-1", 1, 0, false, false, false, false),
		{
			Name:    "gaia",
			Version: "v13.0.0", // TODO: Can we use "latest" here?
			ChainConfig: ibc.ChainConfig{
				ModifyGenesis: modifyGaiaGenesis,
			},
		},
	})

	chains, err := factory.Chains(t.Name())
	require.NoError(t, err)

	noble := chains[0].(*cosmos.CosmosChain)
	gaia := chains[1].(*cosmos.CosmosChain)

	wrapper.chain = noble

	rly := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		logger,
		relayerImage,
	).Build(t, client, network)

	interchain := interchaintest.NewInterchain().
		AddChain(noble).
		AddChain(gaia).
		AddRelayer(rly, "rly").
		AddLink(interchaintest.InterchainLink{
			Chain1:  noble,
			Chain2:  gaia,
			Relayer: rly,
			Path:    "noble-gaia",
		})

	require.NoError(t, interchain.Build(ctx, execReporter, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
	}))

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	// Register Noble as a consumer chain.
	proposers := interchaintest.GetAndFundTestUsers(t, ctx, "proposer", 100_000_000, gaia)
	proposer := proposers[0]

	content := providerTypes.ConsumerAdditionProposal{
		Title:                             "Migrate Noble to ICS",
		Description:                       "This proposal aims to migrate Noble from a standalone POA chain to ICS.",
		ChainId:                           "grand-1",
		InitialHeight:                     ibcclienttypes.NewHeight(1, 75),
		GenesisHash:                       []byte("genesis"),
		BinaryHash:                        []byte("binary"),
		SpawnTime:                         time.Now(),
		UnbondingPeriod:                   time.Hour * 24 * 7 * 3, // 1814400s
		CcvTimeoutPeriod:                  time.Hour * 24 * 7 * 4, // 2419200s
		TransferTimeoutPeriod:             time.Hour,              // 3600s
		ConsumerRedistributionFraction:    "0.85",
		BlocksPerDistributionTransmission: 1000,
		HistoricalEntries:                 10_000,
		//DistributionTransmissionChannel:   "",
	}

	contentBz, err := json.Marshal(ConsumerAdditionJSON{
		ConsumerAdditionProposal: content,
		Deposit:                  "10000000uatom",
	})
	require.NoError(t, err)

	require.NoError(t, gaia.FullNodes[0].WriteFile(ctx, contentBz, "proposal.json"))
	_, err = gaia.FullNodes[0].ExecTx(ctx, proposer.KeyName(), "gov", "submit-proposal", "consumer-addition", filepath.Join(gaia.FullNodes[0].HomeDir(), "proposal.json"), "--gas", "250000")
	require.NoError(t, err)

	require.NoError(t, gaia.VoteOnProposalAllValidators(ctx, "1", "yes"))

	// Register Krypton upgrade on Noble.
	_, err = noble.Validators[0].ExecTx(ctx, wrapper.paramAuthority.KeyName(), "upgrade", "software-upgrade", "krypton", "--upgrade-height", "75")
	require.NoError(t, err)

	current, err := noble.Height(ctx)
	require.NoError(t, err)

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer timeoutCtxCancel()

	_ = testutil.WaitForBlocks(timeoutCtx, int(75-current)+1, noble)

	fmt.Println(noble.Height(ctx))

	require.NoError(t, noble.StopAllNodes(ctx))
	noble.UpgradeVersion(ctx, client, "ghcr.io/strangelove-ventures/noble", "john-test")
	require.NoError(t, noble.StartAllNodes(ctx))

	rly.CreateChannel(ctx, execReporter, "noble-gaia", ibc.CreateChannelOptions{
		SourcePortName: "consumer",
		DestPortName:   "provider",
		Order:          ibc.Ordered,
		Version:        "ccv-1",
	})
}
