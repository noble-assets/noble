package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/require"
)

// circle testnet values
func TestMarshalBurnMessage(t *testing.T) {
	const depositForBurn = "0000000000000000000000000000000007865c6e87b9f70255377e024ace6630c1eaa37f0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a3300000000000000000000000000000000000000000000000000000000007a12000000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	depositForBurnBz, err := hex.DecodeString(depositForBurn)
	require.NoError(t, err)

	bm := new(types.BurnMessage)
	err = bm.UnmarshalBytes(depositForBurnBz)
	require.NoError(t, err)

	require.Equal(t, uint32(0), bm.Version)

	const expectedBurnToken = "00000000000000000000000007865c6e87b9f70255377e024ace6630c1eaa37f"
	expectedBurnTokenBz, err := hex.DecodeString(expectedBurnToken)
	require.NoError(t, err)
	require.Equal(t, expectedBurnTokenBz, bm.BurnToken)

	const expectedMintRecipient = "0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	expectedMintRecipientBz, err := hex.DecodeString(expectedMintRecipient)
	require.NoError(t, err)
	require.Equal(t, expectedMintRecipientBz, bm.MintRecipient)

	require.Equal(t, uint64(0x7a1200), bm.Amount)

	const expectedMessageSender = "0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	expectedMessageSenderBz, err := hex.DecodeString(expectedMessageSender)
	require.NoError(t, err)
	require.Equal(t, expectedMessageSenderBz, bm.MessageSender)

	mintRecipient, err := hex.DecodeString("0000000000000000000000009481EF9E2CA814FC94676DEA3E8C3097B06B3A33")
	require.NoError(t, err)
	require.Equal(t, mintRecipient, bm.MintRecipient)

	marshaledBurn := bm.Bytes()
	require.Equal(t, depositForBurnBz, marshaledBurn)
}
