package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/require"
)

// circle testnet values
func TestMarshalMessage(t *testing.T) {
	// Prepare test data
	// depositForBurn from https://goerli.etherscan.io/tx/0x7323d32ae63a475fdfc2ea75e8ec1cb48113320cc4c107d63a0d046b29c445e5#eventlog
	message, err := hex.DecodeString("0000000000000000000000030000000000039193000000000000000000000000D0C3DA58F55358142B8D3E06C1C30C5C6114EFE800000000000000000000000012DCFD3FE2E9EAC2859FD1ED86D2AB8C5A2F935200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007865C6E87B9F70255377E024ACE6630C1EAA37F0000000000000000000000009481EF9E2CA814FC94676DEA3E8C3097B06B3A3300000000000000000000000000000000000000000000000000000000007A12000000000000000000000000009481EF9E2CA814FC94676DEA3E8C3097B06B3A33")
	require.NoError(t, err)

	m := new(types.Message)
	err = m.UnmarshalBytes(message)
	require.NoError(t, err)

	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, m.DestinationCaller)

	require.Equal(t, uint32(0), m.SourceDomain)

	require.Equal(t, uint64(0x39193), m.Nonce)

	require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x12, 0xdc, 0xfd, 0x3f, 0xe2, 0xe9, 0xea, 0xc2, 0x85, 0x9f, 0xd1, 0xed, 0x86, 0xd2, 0xab, 0x8c, 0x5a, 0x2f, 0x93, 0x52}, m.Recipient)

	require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xd0, 0xc3, 0xda, 0x58, 0xf5, 0x53, 0x58, 0x14, 0x2b, 0x8d, 0x3e, 0x6, 0xc1, 0xc3, 0xc, 0x5c, 0x61, 0x14, 0xef, 0xe8}, m.Sender)

	require.Equal(t, uint32(0), m.Version)

	require.Equal(t, uint32(3), m.DestinationDomain)

	const expectedDepositForBurn = "0000000000000000000000000000000007865c6e87b9f70255377e024ace6630c1eaa37f0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a3300000000000000000000000000000000000000000000000000000000007a12000000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	expectedDepositForBurnBz, err := hex.DecodeString(expectedDepositForBurn)
	require.NoError(t, err)
	require.Equal(t, expectedDepositForBurnBz, m.MessageBody)

	marshaledMsg := m.Bytes()
	require.Equal(t, message, marshaledMsg)
}
