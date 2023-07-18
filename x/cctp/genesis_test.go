package cctp_test

import (
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		Authority: &types.Authority{
			Address: "123",
		},
		PublicKeysList: []types.PublicKeys{
			{
				Key: "0",
			},
			{
				Key: "1",
			},
		},
		MinterAllowanceList: []types.MinterAllowances{
			{
				Denom:  "denom1",
				Amount: 1,
			},
			{
				Denom:  "denom2",
				Amount: 2,
			},
		},
		PerMessageBurnLimit: &types.PerMessageBurnLimit{
			Amount: 23,
		},
		BurningAndMintingPaused: &types.BurningAndMintingPaused{
			Paused: true,
		},
		SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{
			Paused: false,
		},
		MaxMessageBodySize: &types.MaxMessageBodySize{
			Amount: 34,
		},
		Nonce: &types.Nonce{
			Nonce: 34,
		},
		SignatureThreshold: &types.SignatureThreshold{
			Amount: 2,
		},
		TokenPairList: []types.TokenPairs{
			{
				RemoteDomain: uint32(0),
				RemoteToken:  "1",
				LocalToken:   "uusdc",
			},
			{
				RemoteDomain: uint32(1),
				RemoteToken:  "2",
				LocalToken:   "uusdc",
			},
		},
		UsedNoncesList: []types.Nonce{
			{
				Nonce: uint64(1234),
			},
			{
				Nonce: uint64(5678),
			},
		},
	}

	k, ctx := keepertest.CctpKeeper(t)
	cctp.InitGenesis(ctx, k, genesisState)
	got := cctp.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.PublicKeysList, got.PublicKeysList)
	require.Equal(t, genesisState.Authority, got.Authority)
	require.ElementsMatch(t, genesisState.MinterAllowanceList, got.MinterAllowanceList)
	require.Equal(t, genesisState.PerMessageBurnLimit, got.PerMessageBurnLimit)
	require.Equal(t, genesisState.BurningAndMintingPaused, got.BurningAndMintingPaused)
	require.Equal(t, genesisState.SendingAndReceivingMessagesPaused, got.SendingAndReceivingMessagesPaused)
	require.Equal(t, genesisState.MaxMessageBodySize, got.MaxMessageBodySize)
}
