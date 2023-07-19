package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/strangelove-ventures/noble/testutil/keeper"
	"github.com/strangelove-ventures/noble/testutil/nullify"
	"github.com/strangelove-ventures/noble/x/cctp/types"
)

func TestSignatureThresholdGet(t *testing.T) {

	keeper, ctx := keepertest.CctpKeeper(t)

	SignatureThreshold := types.SignatureThreshold{Amount: 2}
	keeper.SetSignatureThreshold(ctx, SignatureThreshold)

	rst, found := keeper.GetSignatureThreshold(ctx)
	require.True(t, found)
	require.Equal(t,
		SignatureThreshold,
		nullify.Fill(&rst),
	)

	newSignatureThreshold := types.SignatureThreshold{Amount: 3}

	keeper.SetSignatureThreshold(ctx, newSignatureThreshold)

	rst, found = keeper.GetSignatureThreshold(ctx)
	require.True(t, found)
	require.Equal(t,
		newSignatureThreshold,
		nullify.Fill(&rst),
	)
}
