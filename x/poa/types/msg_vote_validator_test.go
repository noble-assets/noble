package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// test ValidateBasic for MsgDelegate
func TestMsgVoteValidator(t *testing.T) {
	tests := []struct {
		name       string
		candidate  sdk.ValAddress
		voter      sdk.ValAddress
		inFavor    bool
		expectPass bool
	}{
		{"basic good", valAddr1, valAddr1, true, true},
		{"infavor", valAddr1, valAddr1, true, true},
		{"empty name", nil, valAddr1, true, false},
		{"empty voter", valAddr1, emptyAddr, true, false},
	}

	for _, tc := range tests {
		msg := &MsgVoteValidator{
			VoterAddress:     tc.voter.String(),
			CandidateAddress: tc.candidate.String(),
			InFavor:          tc.inFavor,
		}
		if tc.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", tc.name)
		} else {
			require.NotNil(t, msg.ValidateBasic(), "test: %v", tc.name)
		}
	}
}
