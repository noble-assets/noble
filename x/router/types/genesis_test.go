package types_test

import (
	"testing"

	"github.com/strangelove-ventures/noble/x/router/types"

	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				InFlightPackets: []types.InFlightPacket{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
				Mints: []types.Mint{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
				IbcForwards: []types.StoreIBCForwardMetadata{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
			},
			valid: true,
		},
		{
			desc: "duplicated mints",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				InFlightPackets: []types.InFlightPacket{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
				Mints: []types.Mint{
					{SourceDomainSender: "1"},
					{SourceDomainSender: "1"},
				},
				IbcForwards: []types.StoreIBCForwardMetadata{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated in flight packets",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				InFlightPackets: []types.InFlightPacket{
					{SourceDomainSender: "1"},
					{SourceDomainSender: "1"},
				},
				Mints: []types.Mint{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
				IbcForwards: []types.StoreIBCForwardMetadata{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated ibc forwards",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				InFlightPackets: []types.InFlightPacket{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
				Mints: []types.Mint{
					{SourceDomainSender: "0"},
					{SourceDomainSender: "1"},
				},
				IbcForwards: []types.StoreIBCForwardMetadata{
					{SourceDomainSender: "1"},
					{SourceDomainSender: "1"},
				},
			},
			valid: false,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
