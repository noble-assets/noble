package interchaintest_test

import (
	"context"
	"encoding/json"
	"testing"

	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	globalfeetypes "github.com/noble-assets/noble/v6/x/globalfee/types"
	"github.com/strangelove-ventures/interchaintest/v4/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v4/ibc"
	"github.com/stretchr/testify/require"
)

func testPostRadonUpgrade(
	t *testing.T,
	ctx context.Context,
	noble *cosmos.CosmosChain,
	paramAuthority ibc.Wallet,
) {
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
	require.Equal(t, paramAuthority.FormattedAddress(), distributionEntities[0].Address)
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
