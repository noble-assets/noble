package router_test

import (
	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/router"
	"github.com/strangelove-ventures/noble/x/router/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		InFlightPackets: []types.InFlightPacket{
			{
				SourceDomainSender: "1",
			},
			{
				SourceDomainSender: "2",
			},
		},
		Mints: []types.Mint{
			{
				SourceDomainSender: "3",
			},
			{
				SourceDomainSender: "4",
			},
		},
		IbcForwards: []types.StoreIBCForwardMetadata{
			{
				SourceDomainSender: "5",
			},
			{
				SourceDomainSender: "6",
			},
		},
	}

	k, ctx := keepertest.RouterKeeper(t)
	router.InitGenesis(ctx, k, genesisState)
	got := router.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.InFlightPackets, got.InFlightPackets)
	require.ElementsMatch(t, genesisState.Mints, got.Mints)
	require.ElementsMatch(t, genesisState.IbcForwards, got.IbcForwards)
}
