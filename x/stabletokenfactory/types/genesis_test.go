package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
<<<<<<< HEAD
	"github.com/strangelove-ventures/noble/v4/testutil/sample"
	"github.com/strangelove-ventures/noble/v4/x/stabletokenfactory/types"
=======
	"github.com/noble-assets/noble/v5/testutil/sample"
	"github.com/noble-assets/noble/v5/x/stabletokenfactory/types"
>>>>>>> a4ad980 (chore: rename module path (#283))

	"github.com/stretchr/testify/require"
)

var testAddress = sample.AccAddress()

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
						AddressBz: sample.AddressBz(),
					},
					{
						AddressBz: sample.AddressBz(),
					},
				},
				Paused: &types.Paused{
					Paused: true,
				},
				MasterMinter: &types.MasterMinter{
					Address: sample.AccAddress(),
				},
				MintersList: []types.Minters{
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.NewCoin("test", sdk.NewInt(1)),
					},
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.NewCoin("test", sdk.NewInt(1)),
					},
				},
				Pauser: &types.Pauser{
					Address: sample.AccAddress(),
				},
				Blacklister: &types.Blacklister{
					Address: sample.AccAddress(),
				},
				Owner: &types.Owner{
					Address: sample.AccAddress(),
				},
				MinterControllerList: []types.MinterController{
					{
						Controller: sample.AccAddress(),
						Minter:     sample.AccAddress(),
					},
					{
						Controller: sample.AccAddress(),
						Minter:     sample.AccAddress(),
					},
				},
				MintingDenom: &types.MintingDenom{
					Denom: "test",
				},
			},
			valid: true,
		},
		{
			desc: "invalid privilege separation",
			genState: &types.GenesisState{
				BlacklistedList: []types.Blacklisted{
					{
						AddressBz: sample.AddressBz(),
					},
					{
						sample.AddressBz(),
					},
				},
				Paused: &types.Paused{
					Paused: true,
				},
				MasterMinter: &types.MasterMinter{
					Address: testAddress,
				},
				MintersList: []types.Minters{
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.NewCoin("test", sdk.NewInt(1)),
					},
					{
						Address:   sample.AccAddress(),
						Allowance: sdk.NewCoin("test", sdk.NewInt(1)),
					},
				},
				Pauser: &types.Pauser{
					Address: testAddress,
				},
				Blacklister: &types.Blacklister{
					Address: testAddress,
				},
				Owner: &types.Owner{
					Address: testAddress,
				},
				MinterControllerList: []types.MinterController{
					{
						Controller: sample.AccAddress(),
						Minter:     sample.AccAddress(),
					},
					{
						Controller: sample.AccAddress(),
						Minter:     sample.AccAddress(),
					},
				},
				MintingDenom: &types.MintingDenom{
					Denom: "test",
				},
			},
			valid: false,
		},
		{
			desc: "duplicated blacklisted",
			genState: &types.GenesisState{
				BlacklistedList: []types.Blacklisted{
					{
						AddressBz: []byte("0"),
					},
					{
						AddressBz: []byte("0"),
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
		{
			desc: "duplicated minterController",
			genState: &types.GenesisState{
				MinterControllerList: []types.MinterController{
					{
						Minter: "0",
					},
					{
						Minter: "0",
					},
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
