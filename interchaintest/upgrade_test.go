package interchaintest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkupgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v4"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/strangelove-ventures/interchaintest/v4/testreporter"
	"github.com/strangelove-ventures/interchaintest/v4/testutil"
	"github.com/strangelove-ventures/noble/cmd"
	integration "github.com/strangelove-ventures/noble/interchaintest"
	fiattokenfactorytypes "github.com/strangelove-ventures/noble/x/fiattokenfactory/types"
	globalfeetypes "github.com/strangelove-ventures/noble/x/globalfee/types"
	upgradetypes "github.com/strangelove-ventures/paramauthority/x/upgrade/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

const (
	haltHeightDelta    = uint64(10) // will propose upgrade this many blocks in the future
	blocksAfterUpgrade = uint64(10)
)

type ParamsQueryResponse struct {
	Subspace string `json:"subspace"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

func TestNobleChainUpgrade(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	ctx := context.Background()

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)

	repo, version := integration.GetDockerImageInfo()

	var noble *cosmos.CosmosChain
	var roles NobleRoles
	var paramauthorityWallet Authority

	var (
		upgradeName       = "neon"
		preUpgradeRepo    = "ghcr.io/strangelove-ventures/noble"
		preUpgradeVersion = "v1.0.0"
	)

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
				Repository: preUpgradeRepo,
				Version:    preUpgradeVersion,
				UidGid:     "1025:1025",
			},
		},
		EncodingConfig: NobleEncoding(),
		PreGenesis: func(cc ibc.ChainConfig) error {
			val := noble.Validators[0]
			err := createTokenfactoryRoles(ctx, &roles, DenomMetadata_rupee, val, true)
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
			if err := modifyGenesisTokenfactory(g, "tokenfactory", DenomMetadata_rupee, &roles, true); err != nil {
				return nil, err
			}
			if err := modifyGenesisParamAuthority(g, paramauthorityWallet.Authority.Address); err != nil {
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

	logger := zaptest.NewLogger(t)

	cf := interchaintest.NewBuiltinChainFactory(logger, []*interchaintest.ChainSpec{
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
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	height, err := noble.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight := height + haltHeightDelta

	cmd.SetPrefixes(chainCfg.Bech32Prefix)

	broadcaster := cosmos.NewBroadcaster(t, noble)

	upgradePlan := sdkupgradetypes.Plan{
		Name:   upgradeName,
		Height: int64(haltHeight),
		Info:   upgradeName + " chain upgrade",
	}

	decoded := sdk.MustAccAddressFromBech32(paramauthorityWallet.Authority.Address)
	wallet := &ibc.Wallet{
		Address:  string(decoded),
		Mnemonic: paramauthorityWallet.Authority.Mnemonic,
		KeyName:  paramauthorityWallet.Authority.KeyName,
		CoinType: paramauthorityWallet.Authority.CoinType,
	}

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		wallet,
		&upgradetypes.MsgSoftwareUpgrade{
			Authority: paramauthorityWallet.Authority.Address,
			Plan:      upgradePlan,
		},
	)
	require.NoError(t, err, "error submitting software upgrade tx")

	stdout, stderr, err := noble.Validators[0].ExecQuery(ctx, "upgrade", "plan")
	require.NoError(t, err, "error submitting software upgrade tx")

	logger.Debug("Upgrade", zap.String("plan_stdout", string(stdout)), zap.String("plan_stderr", string(stderr)))

	timeoutCtx, timeoutCtxCancel := context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, noble)

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")

	// bring down nodes to prepare for upgrade
	err = noble.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	noble.UpgradeVersion(ctx, client, "v2.0.0")

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	err = noble.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	haltHeight = height + haltHeightDelta

	upgradeName = "radon"

	upgradePlan = sdkupgradetypes.Plan{
		Name:   upgradeName,
		Height: int64(haltHeight),
		Info:   upgradeName + " chain upgrade",
	}

	_, err = cosmos.BroadcastTx(
		ctx,
		broadcaster,
		wallet,
		&upgradetypes.MsgSoftwareUpgrade{
			Authority: paramauthorityWallet.Authority.Address,
			Plan:      upgradePlan,
		},
	)
	require.NoError(t, err, "error submitting software upgrade tx")

	stdout, stderr, err = noble.Validators[0].ExecQuery(ctx, "upgrade", "plan")
	require.NoError(t, err, "error submitting software upgrade tx")

	logger.Debug("Upgrade", zap.String("plan_stdout", string(stdout)), zap.String("plan_stderr", string(stderr)))

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	// this should timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, noble)

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")

	// make sure that chain is halted
	require.Equal(t, haltHeight, height, "height is not equal to halt height")

	// bring down nodes to prepare for upgrade
	err = noble.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// upgrade version and repo on all nodes
	for _, n := range noble.Nodes() {
		n.Image.Repository = repo
	}
	noble.UpgradeVersion(ctx, client, version)

	// start all nodes back up.
	// validators reach consensus on first block after upgrade height
	// and chain block production resumes.
	err = noble.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, timeoutCtxCancel = context.WithTimeout(ctx, time.Second*45)
	defer timeoutCtxCancel()

	err = testutil.WaitForBlocks(timeoutCtx, int(blocksAfterUpgrade), noble)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	height, err = noble.Height(ctx)
	require.NoError(t, err, "error fetching height after upgrade")

	require.GreaterOrEqual(t, height, haltHeight+blocksAfterUpgrade, "height did not increment enough after upgrade")

	queryResult, _, err := noble.Validators[0].ExecQuery(ctx, "globalfee", "parameters")
	require.NoError(t, err, "error querying globalfee params")

	var globalFeeParams globalfeetypes.Params
	err = json.Unmarshal(queryResult, &globalFeeParams)
	require.NoError(t, err, "failed to unmarshall globalfee params")

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "fiat-tokenfactory", "show-minting-denom")
	require.NoError(t, err, "error querying minting denom")

	var mintingDenomResponse fiattokenfactorytypes.QueryGetMintingDenomResponse
	err = json.Unmarshal(queryResult, &mintingDenomResponse)
	require.NoError(t, err, "failed to unmarshall globalfee params")

	expectedMinGasPrices := sdk.DecCoins{
		sdk.NewDecCoinFromDec(mintingDenomResponse.MintingDenom.Denom, sdk.NewDec(0)),
	}
	require.Equal(t, expectedMinGasPrices, globalFeeParams.MinimumGasPrices, "global fee min gas prices are not as expected")

	require.Equal(t, globalfeetypes.DefaultParams().BypassMinFeeMsgTypes, globalFeeParams.BypassMinFeeMsgTypes, "global fee bypass message types are not as expected")

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "params", "subspace", "tariff", "Share")
	require.NoError(t, err, "error querying tariff 'Share' param")

	var tariffParamShare ParamsQueryResponse

	err = json.Unmarshal(queryResult, &tariffParamShare)
	require.NoError(t, err, "failed to unmarshall tariff share param")

	require.Equal(t, `"`+sdk.NewDecWithPrec(8, 1).String()+`"`, tariffParamShare.Value)

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "params", "subspace", "tariff", "DistributionEntities")
	require.NoError(t, err, "error querying tariff 'DistributionEntities' param")

	var tariffParamDistributionentities ParamsQueryResponse

	err = json.Unmarshal(queryResult, &tariffParamDistributionentities)
	require.NoError(t, err, "failed to unmarshall tariff DistributionEntities param")

	var distributionEntities []DistributionEntity

	err = json.Unmarshal([]byte(tariffParamDistributionentities.Value), &distributionEntities)
	require.NoError(t, err, "failed to unmarshall tariff distribution_entities param")
	require.Len(t, distributionEntities, 1)
	require.Equal(t, paramauthorityWallet.Authority.Address, distributionEntities[0].Address)
	require.Equal(t, sdk.OneDec().String(), distributionEntities[0].Share)
	require.Equal(t, `"`+sdk.NewDecWithPrec(8, 1).String()+`"`, tariffParamShare.Value)

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "params", "subspace", "tariff", "TransferFeeBPS")
	require.NoError(t, err, "failed to unmarshall tariff TransferFeeBPS param")

	var tariffParamTransferFeeBPS ParamsQueryResponse

	err = json.Unmarshal(queryResult, &tariffParamTransferFeeBPS)
	require.NoError(t, err, "failed to unmarshall tariff transfer fee BPS param")

	require.Equal(t, `"`+sdk.OneInt().String()+`"`, tariffParamTransferFeeBPS.Value)

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "params", "subspace", "tariff", "TransferFeeMax")
	require.NoError(t, err, "failed to unmarshall tariff TransferFeeMax param")

	var tariffParamTransferFeeMax ParamsQueryResponse

	err = json.Unmarshal(queryResult, &tariffParamTransferFeeMax)
	require.NoError(t, err, "failed to unmarshall tariff transfer fee BPS param")

	require.Equal(t, `"`+sdk.NewInt(5000000).String()+`"`, tariffParamTransferFeeMax.Value)

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "params", "subspace", "tariff", "TransferFeeDenom")
	require.NoError(t, err, "failed to unmarshall tariff TransferFeeDenom param")

	var tariffParamTransferFeeDenom ParamsQueryResponse

	err = json.Unmarshal(queryResult, &tariffParamTransferFeeDenom)
	require.NoError(t, err, "failed to unmarshall tariff transfer fee BPS param")

	queryResult, _, err = noble.Validators[0].ExecQuery(ctx, "fiat-tokenfactory", "show-minting-denom")
	require.NoError(t, err, "failed to query minting denom")
	var mintingDenom fiattokenfactorytypes.QueryGetMintingDenomResponse

	err = json.Unmarshal(queryResult, &mintingDenom)
	require.NoError(t, err, "failed to unmarshall minting denom")

	require.Equal(t, `"`+mintingDenom.MintingDenom.Denom+`"`, tariffParamTransferFeeDenom.Value)
}
