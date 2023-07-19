package types_test

import (
	"github.com/strangelove-ventures/noble/testutil/sample"
)

var testAddress = sample.AccAddress()

//func TestGenesisState_Validate(t *testing.T) {
//	for _, tc := range []struct {
//		desc     string
//		genState *types.GenesisState
//		valid    bool
//	}{
//		{
//			desc:     "default is invalid",
//			genState: types.DefaultGenesis(),
//			valid:    true,
//		},
//		{
//			desc: "valid genesis state",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: true,
//		},
//		{
//			desc: "authority not set",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "authority is not a valid address",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: "not-an-address",
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "nil public keys list",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "duplicated public keys",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "5"},
//					{Key: "5"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "nil minter allowance list",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "5"},
//					{Key: "4"},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "duplicated minter allowance",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "5"},
//					{Key: "3"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "0", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "per message burn limit is nil",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "max message body size is nil",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//			},
//			valid: false,
//		},
//		{
//			desc: "BurningAndMintingPaused is nil",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//				MaxMessageBodySize:                &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "SendingAndReceivingMessagesPaused is nil",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:     &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused: &types.BurningAndMintingPaused{Paused: true},
//				MaxMessageBodySize:      &types.MaxMessageBodySize{Amount: 34},
//			},
//			valid: false,
//		},
//		{
//			desc: "max message body size is nil",
//			genState: &types.GenesisState{
//				Params: types.DefaultParams(),
//				Authority: &types.Authority{
//					Address: sample.AccAddress(),
//				},
//				PublicKeysList: []types.PublicKeys{
//					{Key: "0"},
//					{Key: "1"},
//				},
//				MinterAllowanceList: []types.MinterAllowances{
//					{Denom: "0", Amount: 123},
//					{Denom: "1", Amount: 456},
//				},
//				PerMessageBurnLimit:               &types.PerMessageBurnLimit{Amount: 23},
//				BurningAndMintingPaused:           &types.BurningAndMintingPaused{Paused: true},
//				SendingAndReceivingMessagesPaused: &types.SendingAndReceivingMessagesPaused{Paused: false},
//			},
//			valid: false,
//		},
//	} {
//		t.Run(tc.desc, func(t *testing.T) {
//			err := tc.genState.Validate()
//			if tc.valid {
//				require.NoError(t, err)
//			} else {
//				require.Error(t, err)
//			}
//		})
//	}
//}
