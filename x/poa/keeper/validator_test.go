package keeper

import (
	"testing"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/protobuf/proto"
	"github.com/strangelove-ventures/noble/x/poa/types"
	"github.com/stretchr/testify/require"
)

func TestKeeperValidatorFunctions(t *testing.T) {
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

	// Set a validator in the store
	keeper.SaveValidator(ctx, validator)

	// Check the store to see if the validator was saved
	retVal, found := keeper.GetValidator(ctx, validator.Address)
	require.True(t, proto.Equal(validator, retVal), "Should return the correct validator from the store")
	require.True(t, found)

	// Get all validators from the store
	allVals := keeper.GetAllValidators(ctx)
	require.Equal(t, 1, len(allVals))
}
