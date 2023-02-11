package types

import (
	"testing"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	pk1      = ed25519.GenPrivKey().PubKey()
	addr1    = pk1.Address()
	valAddr1 = sdk.ValAddress(addr1)

	emptyAddr   sdk.ValAddress
	emptyPubkey cryptotypes.PubKey
)

func TestMsgCreateValidator(t *testing.T) {
	tests := []struct {
		name, moniker, identity, website, securityContact, details string
		validatorAddr                                              sdk.ValAddress
		pubkey                                                     cryptotypes.PubKey
		expectPass                                                 bool
	}{
		{"basic good", "a", "b", "c", "d", "e", valAddr1, pk1, true},
		{"empty name", "", "", "c", "", "", valAddr1, pk1, false},
		{"empty description", "", "", "", "", "", valAddr1, pk1, false},
		{"empty address", "a", "b", "c", "d", "e", emptyAddr, pk1, false},
		{"empty pubkey", "a", "b", "c", "d", "e", valAddr1, emptyPubkey, true},
	}

	for _, tc := range tests {
		description := stakingtypes.NewDescription(tc.moniker, tc.identity, tc.website, tc.securityContact, tc.details)
		pk, err := cdctypes.NewAnyWithValue(tc.pubkey)
		require.NoError(t, err)
		msg := &MsgCreateValidator{
			Description: description,
			Address:     tc.validatorAddr.String(),
			Pubkey:      pk,
		}
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}
