package keeper

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/strangelove-ventures/noble/x/poa/types"
)

// IsValSetSorted reports whether valset is sorted.
func IsValSetSorted(data []stakingtypes.Validator, powerReduction sdk.Int) bool {
	n := len(data)
	for i := n - 1; i > 0; i-- {
		if stakingtypes.ValidatorsByVotingPower(data).Less(i, i-1, powerReduction) {
			return false
		}
	}
	return true
}

func TestHistoricalInfo(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	const numVals = 20

	validators := make([]*types.Validator, numVals)

	// Create test validators
	pubKeys := CreateTestPubKeys(numVals)
	for i, pubKey := range pubKeys {
		valAddr := sdk.ValAddress(pubKey.Address().Bytes()) // would not actually be derived from consensus key, but fine for test

		pubKeyAny, err := cdctypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		validators[i] = &types.Validator{
			Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
			Address:     valAddr,
			Pubkey:      pubKeyAny,
		}

		keeper.SaveValidator(ctx, validators[i])
	}

	stakingVals := keeper.GetAllValidatorsStaking(ctx)

	hi := stakingtypes.NewHistoricalInfo(ctx.BlockHeader(), stakingVals, keeper.PowerReduction(ctx))
	keeper.SetHistoricalInfo(ctx, 2, &hi)

	recv, found := keeper.GetHistoricalInfo(ctx, 2)
	require.True(t, found, "HistoricalInfo not found after set")

	hiMarshaled, err := proto.Marshal(&hi)
	require.NoError(t, err)

	recvMarshaled, err := proto.Marshal(&recv)
	require.NoError(t, err)

	require.Equal(t, hiMarshaled, recvMarshaled, "HistoricalInfo not equal")
	require.True(t, IsValSetSorted(recv.Valset, keeper.PowerReduction(ctx)), "HistoricalInfo validators is not sorted")

	keeper.DeleteHistoricalInfo(ctx, 2)

	recv, found = keeper.GetHistoricalInfo(ctx, 2)
	require.False(t, found, "HistoricalInfo found after delete")
	require.Equal(t, stakingtypes.HistoricalInfo{}, recv, "HistoricalInfo is not empty")
}

func TestTrackHistoricalInfo(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	const numVals = 4

	validators := make([]*types.Validator, numVals)

	// Create test validators
	pubKeys := CreateTestPubKeys(numVals)
	for i, pubKey := range pubKeys {
		valAddr := sdk.ValAddress(pubKey.Address().Bytes()) // would not actually be derived from consensus key, but fine for test

		pubKeyAny, err := cdctypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		validators[i] = &types.Validator{
			Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
			Address:     valAddr,
			Pubkey:      pubKeyAny,
		}
	}

	// set historical entries in params to 5
	params := types.DefaultParams()
	params.HistoricalEntries = 5
	keeper.SetParams(ctx, params)

	// set historical info at 5, 4 which should be pruned
	// and check that it has been stored
	h4 := tmproto.Header{
		ChainID: "HelloChain",
		Height:  4,
	}
	h5 := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	valSet := []stakingtypes.Validator{
		validators[0].ToStakingValidator(),
		validators[1].ToStakingValidator(),
	}
	hi4 := stakingtypes.NewHistoricalInfo(h4, valSet, keeper.PowerReduction(ctx))
	hi5 := stakingtypes.NewHistoricalInfo(h5, valSet, keeper.PowerReduction(ctx))
	keeper.SetHistoricalInfo(ctx, 4, &hi4)
	keeper.SetHistoricalInfo(ctx, 5, &hi5)
	recv, found := keeper.GetHistoricalInfo(ctx, 4)
	require.True(t, found)

	hiMarshaled, err := proto.Marshal(&hi4)
	require.NoError(t, err)

	recvMarshaled, err := proto.Marshal(&recv)
	require.NoError(t, err)

	require.Equal(t, hiMarshaled, recvMarshaled)

	recv, found = keeper.GetHistoricalInfo(ctx, 5)
	require.True(t, found)

	hiMarshaled, err = proto.Marshal(&hi5)
	require.NoError(t, err)

	recvMarshaled, err = proto.Marshal(&recv)
	require.NoError(t, err)

	require.Equal(t, hiMarshaled, recvMarshaled)

	// Set bonded validators in keeper
	val1 := validators[2]
	val1.InSet = true // when not in set, consensus power is Zero
	keeper.SaveValidator(ctx, val1)

	val2 := validators[3]
	val2.InSet = true
	keeper.SaveValidator(ctx, val2)

	vals := []stakingtypes.Validator{val2.ToStakingValidator(), val1.ToStakingValidator()}
	require.True(t, IsValSetSorted(vals, keeper.PowerReduction(ctx)), "HistoricalInfo validators is not sorted")

	// Set Header for BeginBlock context
	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  10,
	}
	ctx = ctx.WithBlockHeader(header)

	keeper.TrackHistoricalInfo(ctx)

	// Check HistoricalInfo at height 10 is persisted
	expected := stakingtypes.HistoricalInfo{
		Header: header,
		Valset: vals,
	}
	recv, found = keeper.GetHistoricalInfo(ctx, 10)
	require.True(t, found, "GetHistoricalInfo failed after BeginBlock")

	hiMarshaled, err = proto.Marshal(&expected)
	require.NoError(t, err)

	recvMarshaled, err = proto.Marshal(&recv)
	require.NoError(t, err)

	require.Equal(t, hiMarshaled, recvMarshaled, "GetHistoricalInfo returned unexpected result")

	// Check HistoricalInfo at height 5, 4 is pruned
	recv, found = keeper.GetHistoricalInfo(ctx, 4)
	require.False(t, found, "GetHistoricalInfo did not prune earlier height")
	require.Equal(t, stakingtypes.HistoricalInfo{}, recv, "GetHistoricalInfo at height 4 is not empty after prune")
	recv, found = keeper.GetHistoricalInfo(ctx, 5)
	require.False(t, found, "GetHistoricalInfo did not prune first prune height")
	require.Equal(t, stakingtypes.HistoricalInfo{}, recv, "GetHistoricalInfo at height 5 is not empty after prune")
}

func TestGetAllHistoricalInfo(t *testing.T) {
	ctx, keeper := MakeTestCtxAndKeeper(t)

	const numVals = 2

	validators := make([]*types.Validator, numVals)

	// Create test validators
	pubKeys := CreateTestPubKeys(numVals)
	for i, pubKey := range pubKeys {
		valAddr := sdk.ValAddress(pubKey.Address().Bytes()) // would not actually be derived from consensus key, but fine for test

		pubKeyAny, err := cdctypes.NewAnyWithValue(pubKey)
		require.NoError(t, err)

		validators[i] = &types.Validator{
			Description: stakingtypes.Description{"nil", "nil", "nil", "nil", "nil"},
			Address:     valAddr,
			Pubkey:      pubKeyAny,
		}
	}

	valSet := []stakingtypes.Validator{
		validators[0].ToStakingValidator(),
		validators[1].ToStakingValidator(),
	}

	header1 := tmproto.Header{ChainID: "HelloChain", Height: 10}
	header2 := tmproto.Header{ChainID: "HelloChain", Height: 11}
	header3 := tmproto.Header{ChainID: "HelloChain", Height: 12}

	hist1 := stakingtypes.HistoricalInfo{Header: header1, Valset: valSet}
	hist2 := stakingtypes.HistoricalInfo{Header: header2, Valset: valSet}
	hist3 := stakingtypes.HistoricalInfo{Header: header3, Valset: valSet}

	expHistInfos := []stakingtypes.HistoricalInfo{hist1, hist2, hist3}

	for i, hi := range expHistInfos {
		keeper.SetHistoricalInfo(ctx, int64(10+i), &hi)
	}

	infos := keeper.GetAllHistoricalInfo(ctx)
	require.Equal(t, expHistInfos, infos)
}
