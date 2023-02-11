package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

func TestKeeperGenericFunctions(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	// Create test validator
	pubKeys := CreateTestPubKeys(1)
	valPubKey := pubKeys[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

	pubKeyAny, err := cdctypes.NewAnyWithValue(valPubKey)
	require.NoError(t, err)

	validator := &types.Validator{
		Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
		Address:     valAddr,
		Pubkey:      pubKeyAny,
	}

	// TODO: split into multiple test cases

	// Set a value in the store
	keeper.Set(ctx, valAddr, types.ValidatorsKey, validator)

	// Check the store to see if the item was saved
	_, found := keeper.Get(ctx, valAddr, types.ValidatorsKey, keeper.UnmarshalValidator)
	require.True(t, found)

	// Get all items from the store
	allVals := keeper.GetAll(ctx, types.ValidatorsKey, keeper.UnmarshalValidator)
	require.Equal(t, 1, len(allVals))
}
