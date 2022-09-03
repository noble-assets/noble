package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"noble/x/tokenfactory/types"
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

				BlacklistedList: []types.Blacklisted{
					{
						Address: "0",
					},
					{
						Address: "1",
					},
				},
				Paused: &types.Paused{
					Paused: true,
				},
				MasterMinter: &types.MasterMinter{
					Address: "82",
				},
				MintersList: []types.Minters{
					{
						Address: "0",
					},
					{
						Address: "1",
					},
				},
				Pauser: &types.Pauser{
					Address: "32",
				},
				Blacklister: &types.Blacklister{
					Address: "78",
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated blacklisted",
			genState: &types.GenesisState{
				BlacklistedList: []types.Blacklisted{
					{
						Address: "0",
					},
					{
						Address: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated minters",
			genState: &types.GenesisState{
				MintersList: []types.Minters{
					{
						Address: "0",
					},
					{
						Address: "0",
					},
				},
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
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
