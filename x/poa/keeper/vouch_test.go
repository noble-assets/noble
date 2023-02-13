package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeeperVouchFunctions(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	// Create test vouch
	pubKeys := CreateTestPubKeys(2)
	voucher, candidate := sdk.ValAddress(pubKeys[0].Address().Bytes()), sdk.ValAddress(pubKeys[1].Address().Bytes())

	vouch := &types.Vouch{
		VoucherAddress:   voucher,
		CandidateAddress: candidate,
		InFavor:          true,
	}

	// TODO: split into multiple test cases

	// Set a vouch in the store
	keeper.SetVouch(ctx, vouch)

	// Check the store to see if the vouch was saved
	retVal, found := keeper.GetVouch(ctx, append(candidate, voucher...))
	assert.Equal(t, vouch, retVal, "Should return the correct vouch from the store")
	require.True(t, found)

	// Get all vouches from the store
	allVals := keeper.GetAllVouches(ctx)
	require.Equal(t, 1, len(allVals))

	// Get all vouches for a validator from the store
	vouchesForVal := keeper.GetAllVouchesForValidator(ctx, candidate)
	require.Equal(t, 1, len(vouchesForVal))
}
