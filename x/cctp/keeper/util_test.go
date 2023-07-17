package keeper_test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/strangelove-ventures/noble/x/cctp/keeper"
	"github.com/strangelove-ventures/noble/x/cctp/types"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

/**
 * We must know the private key to sign messages, so we generate them every test and
 * generate attestations (signatures) and attesters (pubkeys) from them.
 * If attestationOverride or attesterOverride are passed in, these are used instead
 */
func TestVerifyAttestationSignatures(t *testing.T) {
	testCases := []struct {
		name               string
		message            []byte
		validAttesters     int
		signatureThreshold uint32
		privateKeys        []ecdsa.PrivateKey // private keys used to generate valid pubkeys
		attesationOverride []byte
		attesterOverride   []types.Attester
		verified           bool // expected verification result
		err                error
	}{
		{
			name:               "happy path",
			message:            []byte("Hello world!"),
			validAttesters:     3,
			signatureThreshold: 1,
			privateKeys:        GeneratePrivateKeys(3),
			verified:           true,
		},
		{
			name:               "invalid attestation length",
			message:            []byte("Hello world!"),
			validAttesters:     3,
			signatureThreshold: 1,
			privateKeys:        GeneratePrivateKeys(3),
			attesationOverride: []byte("fasd"),
			attesterOverride: []types.Attester{
				{
					Attester: "123",
				},
				{
					Attester: "456",
				},
			},
			verified: true,
			err:      types.ErrSignatureVerification,
		},
	}

	assert := assert.New(t)

	for i, tc := range testCases {
		name := fmt.Sprintf("#%d: VerifyAttesationSignature(%s, %d) should be %t", i, string(tc.message), tc.signatureThreshold, tc.verified)
		t.Run(name, func(t *testing.T) {

			if len(tc.privateKeys) < tc.validAttesters {
				panic(fmt.Sprintf("not enough private keys to generate %d valid attesters", tc.validAttesters))
			}

			// Calculate the hash of the message
			hash := crypto.Keccak256(tc.message)

			var attesters []types.Attester
			var attestations []byte

			if tc.attesterOverride != nil {
				// use test case attesters
				attesters = tc.attesterOverride
			} else {
				// Generate N attester public keys (attesters)
				for _, privateKey := range tc.privateKeys {
					pubKeyBytes := elliptic.Marshal(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
					pubKey := hex.EncodeToString(pubKeyBytes)

					attesters = append(attesters, types.Attester{Attester: pubKey})
				}
			}

			if tc.attesationOverride != nil {
				// use test case attestations
				attestations = tc.attesationOverride
			} else {
				// Sign the hash with each private key to get attestation
				for j := 0; j < tc.validAttesters; j++ {
					signature, err := crypto.Sign(hash[:], &tc.privateKeys[j])
					if err != nil {
						t.Fatalf("Failed to sign message: %v", err)
					}
					// assert that the length of the signature is 65 bytes
					assert.Equal(len(signature), crypto.SignatureLength)
					attestations = append(attestations, signature...)
				}
			}

			verified, err := keeper.VerifyAttestationSignatures(tc.message, attestations, attesters, tc.signatureThreshold)

			assert.Equal(tc.verified, verified)
			assert.Equal(tc.err, err)
			if err != nil {
				assert.Equal("signature threshold is 0", err.Error())
			}
		})
	}
}

// GeneratePrivateKeys generates N private keys
func GeneratePrivateKeys(n int) []ecdsa.PrivateKey {
	var privateKeys []ecdsa.PrivateKey
	for i := 0; i < n; i++ {
		privateKey, _ := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
		privateKeys = append(privateKeys, *privateKey)
	}

	privateKeys = SortPrivateKeys(privateKeys)

	return privateKeys
}

// SortPrivateKeys sorts an array of ECDSA private keys based on their uncompressed public keys
func SortPrivateKeys(privateKeys []ecdsa.PrivateKey) []ecdsa.PrivateKey {
	type KeyWithUncompressedPublicKey struct {
		Key    ecdsa.PrivateKey
		PubKey []byte
	}

	keysWithPubKeys := make([]KeyWithUncompressedPublicKey, len(privateKeys))

	for i, privateKey := range privateKeys {
		pubKeyBytes := elliptic.Marshal(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
		pubKey := hex.EncodeToString(pubKeyBytes)
		keysWithPubKeys[i] = KeyWithUncompressedPublicKey{
			Key:    privateKey,
			PubKey: []byte(pubKey),
		}
	}

	sort.SliceStable(keysWithPubKeys, func(i, j int) bool {
		return bytes.Compare(keysWithPubKeys[i].PubKey, keysWithPubKeys[j].PubKey) < 0
	})

	sortedPrivateKeys := make([]ecdsa.PrivateKey, len(privateKeys))
	for i, keyWithPubKey := range keysWithPubKeys {
		sortedPrivateKeys[i] = keyWithPubKey.Key
	}

	return sortedPrivateKeys
}
