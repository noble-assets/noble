package keeper_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/strangelove-ventures/noble/x/cctp/keeper"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/require"
)

// circle testnet values
func TestVerifyAttestationSignatures(t *testing.T) {
	// Prepare test data
	// depositForBurn from https://goerli.etherscan.io/tx/0x7323d32ae63a475fdfc2ea75e8ec1cb48113320cc4c107d63a0d046b29c445e5#eventlog
	message, _ := hex.DecodeString("0000000000000000000000030000000000039193000000000000000000000000D0C3DA58F55358142B8D3E06C1C30C5C6114EFE800000000000000000000000012DCFD3FE2E9EAC2859FD1ED86D2AB8C5A2F935200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007865C6E87B9F70255377E024ACE6630C1EAA37F0000000000000000000000009481EF9E2CA814FC94676DEA3E8C3097B06B3A3300000000000000000000000000000000000000000000000000000000007A12000000000000000000000000009481EF9E2CA814FC94676DEA3E8C3097B06B3A33")

	// https://iris-api-sandbox.circle.com/attestations/0xd4d42eb5e134d3fe8d397e82d610c57b5fc3abae2d3d976532d52981dab0eb2b
	attestation, _ := hex.DecodeString("640692cd7b0332ca6f38196f232752bca1a01619538d7e3ce6183f317c66d2483f4c7e83f2642fca25c384effc23647dec3cda7bc4ceac78c8681a22b7a351891bc5c3fedba209d406ec6cdfb434f62bb15d9623acc81d3769cb3d7987c5cb5d9b76b7ad0161e0985b4ae691c5334697fad1b5196ee104de19c09f9718751a740f1c")

	// from https://iris-api-sandbox.circle.com/v1/publicKeys
	publicKeys := []types.PublicKeys{
		{Key: "048af0d36997eb1775c1a74a539b737f0a8db209ea7fc5677e64055a730ed07ad288fa3e7b3871ff18a4a55c4273c16ac5a2cff30dc00a122f63b3e655a62e093c"},
		{Key: "04a36d8f1818402096aacaf7340493bdea39d1728948c40bbfb711feb7be35960139077dd8332be75f188f8567b4fc0fcf649fddf3397e569c856bf9ce9329d3b9"},
	}

	depositForBurnWrapper := keeper.ParseIntoMessage(message)
	depositForBurn := keeper.ParseIntoBurnMessage(depositForBurnWrapper.MessageBody)
	require.Equal(t, uint32(3), depositForBurnWrapper.DestinationDomain)

	const expectedDepositForBurn = "0000000000000000000000000000000007865c6e87b9f70255377e024ace6630c1eaa37f0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a3300000000000000000000000000000000000000000000000000000000007a12000000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	expectedDepositForBurnBz, err := hex.DecodeString(expectedDepositForBurn)
	require.NoError(t, err)
	require.True(t, bytes.Equal(expectedDepositForBurnBz, depositForBurnWrapper.MessageBody))

	require.Equal(t, uint32(0), depositForBurn.Version)

	const expectedBurnToken = "00000000000000000000000007865c6e87b9f70255377e024ace6630c1eaa37f"
	expectedBurnTokenBz, err := hex.DecodeString(expectedBurnToken)
	require.NoError(t, err)
	require.Equal(t, expectedBurnTokenBz, depositForBurn.BurnToken)

	const expectedMintRecipient = "0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	expectedMintRecipientBz, err := hex.DecodeString(expectedMintRecipient)
	require.NoError(t, err)
	require.Equal(t, expectedMintRecipientBz, depositForBurn.MintRecipient)

	require.Equal(t, uint64(0x7a1200), depositForBurn.Amount)

	const expectedMessageSender = "0000000000000000000000009481ef9e2ca814fc94676dea3e8c3097b06b3a33"
	expectedMessageSenderBz, err := hex.DecodeString(expectedMessageSender)
	require.NoError(t, err)
	require.Equal(t, expectedMessageSenderBz, depositForBurn.MessageSender)

	mintRecipient, err := hex.DecodeString("0000000000000000000000009481EF9E2CA814FC94676DEA3E8C3097B06B3A33")
	require.NoError(t, err)
	require.True(t, bytes.Equal(mintRecipient, depositForBurn.MintRecipient))

	marshaledBurn := keeper.ParseBurnMessageIntoBytes(depositForBurn)
	require.Equal(t, expectedDepositForBurnBz, marshaledBurn)

	marshaledMsg := keeper.ParseIntoMessageBytes(depositForBurnWrapper)
	require.Equal(t, message, marshaledMsg)

	t.Run("valid input", func(t *testing.T) {
		ok, err := keeper.VerifyAttestationSignatures(message, attestation, publicKeys, 2)
		require.NoError(t, err)
		require.True(t, ok)
	})

}

// local values
func TestWithGeneratedValues(t *testing.T) {
	privKey1, _ := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	privKey2, _ := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)

	pubKey1 := crypto.FromECDSAPub(&privKey1.PublicKey)
	pubKey2 := crypto.FromECDSAPub(&privKey2.PublicKey)
	publicKeys := []types.PublicKeys{
		{Key: hex.EncodeToString(pubKey1)},
		{Key: hex.EncodeToString(pubKey2)},
	}

	message := []byte("Hello, World!")
	digest := crypto.Keccak256(message)
	signature, _ := crypto.Sign(digest, privKey1)
	signature2, _ := crypto.Sign(digest, privKey2)

	verified, err := keeper.VerifyAttestationSignatures(message, append(signature, signature2...), publicKeys, 2)
	require.NoError(t, err)
	require.True(t, verified)
}
