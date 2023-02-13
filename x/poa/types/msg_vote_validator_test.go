package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// test ValidateBasic for MsgDelegate
func TestMsgVouchValidator(t *testing.T) {
	tests := []struct {
		name       string
		candidate  sdk.ValAddress
		vouchr     sdk.ValAddress
		inFavor    bool
		expectPass bool
	}{
		{"basic good", valAddr1, valAddr1, true, true},
		{"infavor", valAddr1, valAddr1, true, true},
		{"empty name", nil, valAddr1, true, false},
		{"empty vouchr", valAddr1, emptyAddr, true, false},
	}

	for _, tc := range tests {
		msg := &MsgVouchValidator{
			VoucherAddress:   tc.vouchr.String(),
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
